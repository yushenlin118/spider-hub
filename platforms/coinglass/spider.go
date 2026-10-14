package coinglass

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// CoinglassResponse API响应结构
// 🔐 所有Coinglass API都返回加密的响应数据
// 📦 需要通过特定的解密流程才能获取原始JSON数据
type CoinglassResponse struct {
	Code    string `json:"code"`    // 📋 响应状态码
	Msg     string `json:"msg"`     // 💬 响应消息
	Data    string `json:"data"`    // 🔒 加密的数据内容
	Success bool   `json:"success"` // ✅ 请求是否成功
}

// Spider Coinglass爬虫
// 🕷️ 专门用于抓取Coinglass平台加密货币数据的爬虫实例
// 🔧 内置HTTP客户端和加密解密功能
type Spider struct {
	client *resty.Client // 🌐 HTTP请求客户端

	loginToken  string     // 🔑 登录凭证（浏览器 cookie「obe」的值），为空表示未登录
	userAgent   string     // 🎭 自定义 User-Agent，为空则沿用 RequestHeaders
	mu          sync.Mutex // 🔒 保护 lastRequest，确保并发调用时限速仍然有效
	lastRequest time.Time  // ⏱️ 上一次请求时间，登录状态下用于限速
}

// ErrLoginRequired 服务端要求登录（业务码 40000）
// 🚪 未调用 SetLoginToken，或 obe 已过期/被登出时返回
// 🛡️ 遇到此错误请勿重试，重新从浏览器取得 obe 后再请求，避免触发风控
var ErrLoginRequired = errors.New("❌ 需要登录：obe 未设置或已失效，请从浏览器重新取得")

// NewSpider 创建新的Coinglass爬虫实例
// 🏗️ 初始化爬虫配置，设置超时时间和请求参数
// 📝 返回可用于数据抓取的Spider实例
//
// 使用示例:
//
//	spider := NewSpider()
//	data, err := spider.GetData("https://capi.coinglass.com/api/...")
func NewSpider() *Spider {
	client := resty.New()
	client.SetTimeout(RequestTimeoutSeconds * time.Second) // ⏱️ 设置请求超时

	return &Spider{
		client: client,
	}
}

// SetLoginToken 设置登录凭证，用于访问需要登录的接口（如清算地图）
// 🔑 token 为浏览器登录 coinglass 后 cookie「obe」的值
// 📝 取得方式: DevTools -> Application -> Cookies -> www.coinglass.com -> obe
// 🛡️ 设置后所有请求自动限速（间隔 LoginMinRequestIntervalSeconds 秒），降低账号被封风险
//
// ⚠️ obe 等同账号登录态，请勿写进代码或提交到仓库，建议从环境变量读取
func (s *Spider) SetLoginToken(token string) {
	s.loginToken = strings.TrimSpace(token)
}

// SetUserAgent 设置自定义 User-Agent
// 🎭 建议与取得 obe 的那个浏览器保持一致，同一登录态在不同UA间跳动容易被风控识别
func (s *Spider) SetUserAgent(ua string) {
	s.userAgent = strings.TrimSpace(ua)
}

// waitRateLimit 登录状态下的请求限速
// 🛡️ 未登录时不限速，保持原有行为；登录时确保两次请求至少间隔 LoginMinRequestIntervalSeconds 秒
func (s *Spider) waitRateLimit() {
	if s.loginToken == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	minInterval := LoginMinRequestIntervalSeconds * time.Second
	if wait := minInterval - time.Since(s.lastRequest); wait > 0 {
		fmt.Printf("⏳ 登录状态限速，等待 %.1f 秒\n", wait.Seconds())
		time.Sleep(wait)
	}
	s.lastRequest = time.Now()
}

// GetData 获取任意API数据（通用方法）
// 🌟 这是Spider的核心方法，用于获取任意Coinglass API的数据
// 🔄 自动处理加密请求、数据解密、gzip解压等复杂流程
// 📊 支持所有Coinglass公开API接口
//
// 参数:
//
//	apiURL - 🌐 完整的API请求URL，包含查询参数
//
// 返回:
//
//	string - 📄 解密后的JSON字符串数据
//	error  - ❌ 如果请求或解密过程中出现错误
//
// 支持的API接口:
//
//	📈 /api/openInterest/v3/chart - 持仓量图表数据
//	🏠 /api/futures/home/statistics - 期货首页统计
//	🏢 /api/derivative/exchange/list - 衍生品交易所列表
//	💰 /api/home/v2/coinMarkets - 币种市场数据
//	🔗 /api/exchange/futures/pairInfo - 期货交易对信息
//	🪙 /api/spot/support/coin - 现货支持币种
func (s *Spider) GetData(apiURL string) (string, error) {
	// ⏰ 生成时间戳作为加密密钥的一部分
	cacheTsV2 := fmt.Sprintf("%d", time.Now().UnixMilli())

	// 🔐 第一步：获取加密的响应数据、动态密钥和加密版本标识
	response, meta, err := s.getEncryptedData(apiURL, cacheTsV2)
	if err != nil {
		return "", fmt.Errorf("获取加密数据失败: %v", err)
	}

	// 🚪 业务码 40000 = 需要登录，此时 data 为空，直接返回明确错误而非解密失败
	if response.Code == LoginRequiredCode {
		return "", ErrLoginRequired
	}

	// 🔑 第二步：根据响应头 v 决定密钥来源
	keySource, err := s.deriveKeySource(meta, apiURL, cacheTsV2)
	if err != nil {
		return "", err
	}

	// 🔓 第三步：解密数据并解压gzip
	decryptedData, err := s.decryptData(response, keySource, meta.User)
	if err != nil {
		return "", fmt.Errorf("解密数据失败: %v", err)
	}

	return decryptedData, nil
}

// getEncryptedData 获取加密响应数据
// 🌐 向Coinglass API发送请求，获取加密的响应数据
// 🔑 同时获取用于解密的动态密钥（通过response header传递）
// 🎭 模拟真实浏览器的请求头，避免被反爬虫检测
//
// 参数:
//
//	apiURL    - 🌐 API请求地址
//	cacheTsV2 - ⏰ 时间戳，用于生成解密密钥
//
// 返回:
//
//	*CoinglassResponse - 📦 加密的API响应结构
//	string            - 🔑 用户动态密钥（从response header获取）
//	error             - ❌ 请求错误信息
func (s *Spider) getEncryptedData(apiURL, cacheTsV2 string) (*CoinglassResponse, *ResponseMeta, error) {
	// 🎭 复用constants.go中定义的浏览器请求头，模拟真实用户访问
	headers := make(map[string]string, len(RequestHeaders)+1)
	for k, v := range RequestHeaders {
		headers[k] = v
	}
	headers["cache-ts-v2"] = cacheTsV2 // ⏰ 时间戳密钥，每次请求动态生成
	if s.loginToken != "" {
		headers[LoginHeaderName] = s.loginToken // 🔑 登录凭证，与前端拦截器行为一致
	}
	if s.userAgent != "" {
		headers["user-agent"] = s.userAgent // 🎭 与浏览器登录态保持一致
	}

	// 🛡️ 登录状态下限速
	s.waitRateLimit()

	// 🚀 发送HTTP GET请求
	resp, err := s.client.R().SetHeaders(headers).Get(apiURL)
	if err != nil {
		return nil, nil, fmt.Errorf("请求失败: %v", err)
	}

	// 🔍 检查HTTP状态码
	if resp.StatusCode() != 200 {
		return nil, nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode())
	}

	// 🔑 从响应头获取动态解密密钥与加密版本标识
	meta := &ResponseMeta{
		User: resp.Header().Get("user"),
		V:    resp.Header().Get("v"),
		Time: resp.Header().Get("time"),
	}
	// 📦 解析JSON响应体
	var response CoinglassResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 🚪 需要登录时服务端返回明文 {"code":"40000"} 且不带 user header，
	// 必须先放行给 GetData 判断，否则会被误报为「未找到user header」
	if response.Code == LoginRequiredCode {
		return &response, meta, nil
	}

	if meta.User == "" {
		return nil, nil, fmt.Errorf("未找到user header")
	}

	return &response, meta, nil
}

// decryptData 解密数据
// 🔓 执行完整的数据解密流程，包括多层加密解密和数据解压
// 🔐 使用时间戳密钥解密用户动态密钥
// 🗜️ 使用动态密钥解密响应数据并解压gzip
//
// 解密流程:
//  1. 📅 使用时间戳生成第一层密钥
//  2. 🔑 解密user header获取动态密钥
//  3. 📦 使用动态密钥解密响应数据
//  4. 🗜️ 解压gzip获取最终JSON数据
//
// 参数:
//
//	response   - 📦 加密的API响应
//	cacheTsV2  - ⏰ 时间戳密钥
//	userHeader - 🔑 加密的动态密钥
//
// 返回:
//
//	string - 📄 解密后的JSON字符串
//	error  - ❌ 解密过程中的错误
func (s *Spider) decryptData(response *CoinglassResponse, keySource, userHeader string) (string, error) {
	// 🔑 第一步：密钥来源Base64编码后取前16位（对应前端 btoa(a).substring(0,16)）
	encoded := base64.StdEncoding.EncodeToString([]byte(keySource))
	if len(encoded) < 16 {
		return "", fmt.Errorf("密钥来源过短: %q", keySource)
	}
	timestampKey := []byte(encoded[:16])

	// 🔓 第二步：解密user header获取动态密钥
	userCiphertext, err := base64.StdEncoding.DecodeString(userHeader)
	if err != nil {
		return "", fmt.Errorf("解码user header失败: %v", err)
	}

	// 🔐 使用时间戳密钥解密用户动态密钥
	dynamicKeyBytes, err := s.decryptAES(userCiphertext, timestampKey)
	if err != nil {
		return "", fmt.Errorf("解密user header失败: %v", err)
	}

	// 🔄 转换动态密钥格式并解析
	dynamicZipKey := hex.EncodeToString(dynamicKeyBytes)
	dynamicKey, err := s.parseDecryptedHex(dynamicZipKey)
	if err != nil {
		return "", fmt.Errorf("解析动态密钥失败: %v", err)
	}

	// 🔓 第三步：解密响应数据
	ciphertext, err := base64.StdEncoding.DecodeString(response.Data)
	if err != nil {
		return "", fmt.Errorf("解码响应数据失败: %v", err)
	}

	// 🔐 使用动态密钥解密响应数据
	dynamicKeyBytes = []byte(dynamicKey)
	decryptedData, err := s.decryptAES(ciphertext, dynamicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("解密响应数据失败: %v", err)
	}

	// 🗜️ 第四步：解压gzip获取最终数据
	hexString := hex.EncodeToString(decryptedData)
	result, err := s.parseDecryptedHex(hexString)
	if err != nil {
		return "", fmt.Errorf("解压最终数据失败: %v", err)
	}

	return result, nil
}

// decryptAES 执行AES-ECB解密
// 🔐 使用AES算法的ECB模式进行数据解密
// 🔑 支持128位密钥长度（16字节）
// 📦 自动处理PKCS7填充的去除
//
// AES-ECB特点:
//
//	🔒 每个数据块独立加密，相同明文产生相同密文
//	⚡ 加密速度快，适合大量数据处理
//	🚫 安全性相对较低，但足够用于API数据传输
//
// 参数:
//
//	ciphertext - 🔒 待解密的密文数据
//	key        - 🔑 16字节的AES密钥
//
// 返回:
//
//	[]byte - 📄 解密后的原始数据
//	error  - ❌ 解密过程中的错误
func (s *Spider) decryptAES(ciphertext, key []byte) ([]byte, error) {
	// 🔍 验证密文长度必须是AES块大小的倍数
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度不是块大小的倍数")
	}

	// 🔑 验证密钥长度必须为16字节（AES-128）
	if len(key) != 16 {
		return nil, fmt.Errorf("密钥长度必须为16字节")
	}

	// 🔐 初始化AES解密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("初始化AES失败: %v", err)
	}

	// 🔄 逐块解密数据（ECB模式）
	decrypted := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], ciphertext[i:i+aes.BlockSize])
	}

	// 📦 去除PKCS7填充
	return s.pkcs7Unpad(decrypted)
}

// pkcs7Unpad 去除PKCS7填充
// �� PKCS7是一种标准的数据填充方式，用于确保数据长度符合加密算法要求
// 🔍 填充值等于填充字节的数量，例如填充3个字节则每个字节的值都是3
// ✂️ 解密后需要去除这些填充字节以获取原始数据
//
// PKCS7填充规则:
//
//	📏 如果数据长度刚好是块大小的倍数，仍需添加一个完整的填充块
//	🔢 填充字节的值等于填充的字节数
//	✅ 例如：...ABC + [5,5,5,5,5] 表示填充了5个字节
//
// 参数:
//
//	data - 📄 包含PKCS7填充的解密数据
//
// 返回:
//
//	[]byte - 📄 去除填充后的原始数据
//	error  - ❌ 填充格式错误
func (s *Spider) pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("无效的解密数据长度")
	}

	// 🔢 获取填充长度（最后一个字节的值）
	paddingLen := int(data[length-1])

	// 🔍 验证填充长度的合理性
	if paddingLen > length || paddingLen > aes.BlockSize {
		return nil, fmt.Errorf("无效的PKCS7填充")
	}

	// ✅ 验证填充字节的正确性
	for i := length - paddingLen; i < length; i++ {
		if data[i] != byte(paddingLen) {
			return nil, fmt.Errorf("无效的PKCS7填充字节")
		}
	}

	// ✂️ 返回去除填充后的数据
	return data[:length-paddingLen], nil
}

// parseDecryptedHex 解析十六进制字符串并解压gzip
// 🔄 将十六进制字符串转换为字节数组
// 🗜️ 使用gzip算法解压数据获取最终的JSON字符串
// 📊 Coinglass使用gzip压缩来减少数据传输量
//
// 处理流程:
//  1. 🔤 十六进制字符串 -> 字节数组
//  2. 🗜️ gzip解压 -> 原始JSON字符串
//  3. 📄 返回可解析的JSON数据
//
// 参数:
//
//	hexString - 🔤 十六进制编码的gzip压缩数据
//
// 返回:
//
//	string - 📄 解压后的JSON字符串
//	error  - ❌ 解析或解压过程中的错误
func (s *Spider) parseDecryptedHex(hexString string) (string, error) {
	// 🔤 将十六进制字符串转换为字节数组
	data, err := hex.DecodeString(hexString)
	if err != nil {
		return "", fmt.Errorf("解码十六进制失败: %v", err)
	}

	// 📖 创建字节读取器
	reader := bytes.NewReader(data)

	// 🗜️ 创建gzip解压器
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return "", fmt.Errorf("创建gzip reader失败: %v", err)
	}
	defer gzipReader.Close() // 🔒 确保资源释放

	// 📄 解压数据到缓冲区
	var decompressed bytes.Buffer
	if _, err := io.Copy(&decompressed, gzipReader); err != nil {
		return "", fmt.Errorf("解压gzip失败: %v", err)
	}

	// 📊 返回解压后的JSON字符串
	return decompressed.String(), nil
}

// GetETFFlow 获取比特币现货ETF每日净流入流出
// 🏦 这是前台图表「比特币现货ETF净流入流出 (BTC)」的数据源
// 📊 返回自ETF上市首日至今的完整历史，按日期升序排列
// 💡 Change 字段单位为BTC，ChangeUsd 为美元；每日附带各发行商明细
//
// ⚠️ 该接口不支持分页，单次返回全量数据（约600KB）
//
// 返回:
//
//	[]ETFFlowDay - 📅 按日期升序的每日流入流出数据
//	error        - ❌ 请求或解析失败
//
// 使用示例:
//
//	days, err := spider.GetETFFlow()
//	latest := days[len(days)-1]  // 最新一日
func (s *Spider) GetETFFlow() ([]ETFFlowDay, error) {
	data, err := s.GetData(BaseURL + ETFFlowEndpoint)
	if err != nil {
		return nil, fmt.Errorf("❌ 获取ETF流入流出数据失败: %w", err)
	}

	var days []ETFFlowDay
	if err := json.Unmarshal([]byte(data), &days); err != nil {
		return nil, fmt.Errorf("❌ 解析ETF流入流出数据失败: %w", err)
	}

	return days, nil
}

// GetETFSpotInFlow 获取比特币现货ETF每日净流入及资产快照
// 📉 精简版数据，含当日净流入（美元）、ETF总净资产、BTC价格，但无发行商拆分
// 📊 适合只关心总量趋势的场景
//
// 返回:
//
//	[]ETFSpotInFlowDay - 📅 按日期升序的每日数据
//	error              - ❌ 请求或解析失败
func (s *Spider) GetETFSpotInFlow() ([]ETFSpotInFlowDay, error) {
	data, err := s.GetData(BaseURL + ETFSpotInFlowEndpoint)
	if err != nil {
		return nil, fmt.Errorf("❌ 获取ETF净流入数据失败: %w", err)
	}

	var days []ETFSpotInFlowDay
	if err := json.Unmarshal([]byte(data), &days); err != nil {
		return nil, fmt.Errorf("❌ 解析ETF净流入数据失败: %w", err)
	}

	return days, nil
}

// GetETFList 获取比特币现货ETF清单
// 📋 返回各ETF的基本信息: 代码、发行商、托管方、AUM、管理费率等
//
// 返回:
//
//	[]ETFInfo - 🏦 ETF信息列表
//	error     - ❌ 请求或解析失败
func (s *Spider) GetETFList() ([]ETFInfo, error) {
	data, err := s.GetData(BaseURL + ETFListEndpoint)
	if err != nil {
		return nil, fmt.Errorf("❌ 获取ETF清单失败: %w", err)
	}

	var list []ETFInfo
	if err := json.Unmarshal([]byte(data), &list); err != nil {
		return nil, fmt.Errorf("❌ 解析ETF清单失败: %w", err)
	}

	return list, nil
}

// GetCoinsHeatmap 获取币种热力图数据
// 🔥 这是首页「热力图(24小时)」的数据源，返回全部币种（实测1612个）
// 📊 PriceChangePercent 恒为24小时涨跌幅（决定方块颜色）
// 📐 Value 的含义由 heatmapType 决定（决定方块大小）
//
// 参数:
//
//	heatmapType - 🎨 指标类型，须为 HeatmapTypePrice / HeatmapTypeVol / HeatmapTypeOI 之一
//
// 返回:
//
//	[]HeatmapCoin - 🪙 币种热力图数据
//	error         - ❌ 请求或解析失败
//
// ⚠️ 传入其他值（含数字）服务端会返回 Server Error
//
// 使用示例:
//
//	coins, err := spider.GetCoinsHeatmap(coinglass.HeatmapTypeVol)
//	// coins[i].Value 为该币种24小时成交额（美元）
func (s *Spider) GetCoinsHeatmap(heatmapType string) ([]HeatmapCoin, error) {
	switch heatmapType {
	case HeatmapTypePrice, HeatmapTypeVol, HeatmapTypeOI:
	default:
		return nil, fmt.Errorf("❌ 无效的热力图类型: %q（可选值: %s / %s / %s）",
			heatmapType, HeatmapTypePrice, HeatmapTypeVol, HeatmapTypeOI)
	}

	data, err := s.GetData(BaseURL + CoinsHeatmapEndpoint + "?type=" + heatmapType)
	if err != nil {
		return nil, fmt.Errorf("❌ 获取币种热力图失败: %w", err)
	}

	var coins []HeatmapCoin
	if err := json.Unmarshal([]byte(data), &coins); err != nil {
		return nil, fmt.Errorf("❌ 解析币种热力图失败: %w", err)
	}

	return coins, nil
}

// GetFearGreedIndex 获取恐惧贪婪指数历史
// 😱 对应前台「恐惧&贪婪指数」页（/pro/i/FearGreedIndex），附带当日BTC价格
// 🔓 不需要登录
//
// 参数:
//
//	limit - 📏 取最近 N 笔；< 1 时取全量（实测约 3100+ 笔，自 2018-02-01 起）
//
// 返回:
//
//	[]FearGreedPoint - 📅 按日期升序的指数数据
//	error            - ❌ 请求或解析失败
//
// 使用示例:
//
//	points, err := spider.GetFearGreedIndex(30)
//	latest := points[len(points)-1]  // latest.Value / latest.Level()
func (s *Spider) GetFearGreedIndex(limit int) ([]FearGreedPoint, error) {
	// 📏 size 为空字符串即全量（与前端行为一致）
	size := ""
	if limit >= 1 {
		size = strconv.Itoa(limit)
	}

	data, err := s.GetData(BaseURL + FearGreedIndexEndpoint + "?size=" + size)
	if err != nil {
		return nil, fmt.Errorf("❌ 获取恐惧贪婪指数失败: %w", err)
	}

	return parseFearGreedIndex([]byte(data))
}

// parseFearGreedIndex 将恐惧贪婪指数原始JSON解析为 []FearGreedPoint
// 📝 只取 data[0]（前端同样只用它），三个数组按下标合并
func parseFearGreedIndex(raw []byte) ([]FearGreedPoint, error) {
	var series []struct {
		Dates  []int64    `json:"dates"`
		Values []float64  `json:"values"`
		Prices []*float64 `json:"prices"` // ⚠️ 可能含 null
	}
	if err := json.Unmarshal(raw, &series); err != nil {
		return nil, fmt.Errorf("❌ 解析恐惧贪婪指数失败: %w", err)
	}
	if len(series) == 0 {
		return nil, fmt.Errorf("❌ 恐惧贪婪指数返回为空")
	}

	fg := series[0]
	if len(fg.Values) != len(fg.Dates) || len(fg.Prices) != len(fg.Dates) {
		return nil, fmt.Errorf("❌ 恐惧贪婪指数数组长度不一致: dates=%d values=%d prices=%d",
			len(fg.Dates), len(fg.Values), len(fg.Prices))
	}

	points := make([]FearGreedPoint, len(fg.Dates))
	for i := range fg.Dates {
		points[i] = FearGreedPoint{Date: fg.Dates[i], Value: fg.Values[i]}
		if fg.Prices[i] != nil {
			points[i].Price = *fg.Prices[i]
		}
	}
	return points, nil
}

// GetGlobalM2SupplyGrowth 获取比特币 vs 全球M2货币供应增长
// 💵 对应前台「比特币 vs 全球M2货币供应增长」页（/pro/i/bitcoin-m2-supply-growth）
// 🌐 接口在 FapiBaseURL（fapi.coinglass.com）下，加密协议相同
// 🔓 不需要登录，不支持分页，单次返回全量
//
// 返回:
//
//	[]M2SupplyGrowthPoint - 📅 按日期升序的每周数据
//	error                 - ❌ 请求或解析失败
//
// 使用示例:
//
//	points, err := spider.GetGlobalM2SupplyGrowth()
//	latest := points[len(points)-1]  // latest.M2Supply / latest.YoYGrowth
func (s *Spider) GetGlobalM2SupplyGrowth() ([]M2SupplyGrowthPoint, error) {
	data, err := s.GetData(FapiBaseURL + GlobalM2SupplyGrowthEndpoint)
	if err != nil {
		return nil, fmt.Errorf("❌ 获取全球M2供应增长失败: %w", err)
	}

	var points []M2SupplyGrowthPoint
	if err := json.Unmarshal([]byte(data), &points); err != nil {
		return nil, fmt.Errorf("❌ 解析全球M2供应增长失败: %w", err)
	}

	return points, nil
}

// GetSpotFlow 获取现货资金流入流出
// 🏪 对应前台 /spot-inflow-outflow 顶部图表切到「现货」
// 🔓 不需要登录
//
// 参数:
//
//	symbol   - 🪙 币种，如 BTC
//	interval - ⏱️ 时间粒度，须为 FlowIntervals 的 key，如 h24 / h1 / w1
//	limit    - 📏 笔数；< 1 或超过 FlowMaxLimit 时退回 FlowDefaultLimit
//
// 返回:
//
//	[]FlowBar - 📅 按时间升序的K线
//	error     - ❌ 参数错误、请求或解析失败
//
// 使用示例:
//
//	bars, err := spider.GetSpotFlow("BTC", "h24", 30)
//	latest := bars[len(bars)-1]  // latest.InflowUSD / latest.NetFlowUSD()
func (s *Spider) GetSpotFlow(symbol, interval string, limit int) ([]FlowBar, error) {
	return s.getFlow(FlowMarketSpot, symbol, interval, limit)
}

// GetFuturesFlow 获取合约资金流入流出
// 📜 对应前台 /spot-inflow-outflow 顶部图表切到「合约」，参数与返回同 GetSpotFlow
// 💡 合约成交额远大于现货（BTC 约 10 倍以上）
func (s *Spider) GetFuturesFlow(symbol, interval string, limit int) ([]FlowBar, error) {
	return s.getFlow(FlowMarketFutures, symbol, interval, limit)
}

// getFlow 资金流入流出共用请求逻辑
func (s *Spider) getFlow(market, symbol, interval string, limit int) ([]FlowBar, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("❌ symbol 不能为空")
	}
	if !FlowIntervals[interval] {
		return nil, fmt.Errorf("❌ 无效的时间粒度: %q（可选: m1 / m5 / m15 / m30 / h1 / h2 / h4 / h6 / h8 / h12 / h24 / w1）", interval)
	}
	// 📏 分页参数自行修正，不报错
	if limit < 1 || limit > FlowMaxLimit {
		limit = FlowDefaultLimit
	}

	params := url.Values{}
	params.Set("diff", "false")
	params.Set("minLimit", "false")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("interval", interval)
	params.Set("symbol", fmt.Sprintf(FlowSymbolFormat, symbol, market))

	data, err := s.GetData(BaseURL + FlowKlineEndpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("❌ 获取资金流入流出失败: %w", err)
	}

	var bars []FlowBar
	if err := json.Unmarshal([]byte(data), &bars); err != nil {
		return nil, fmt.Errorf("❌ 解析资金流入流出失败: %w", err)
	}
	return bars, nil
}

// GetLiquidationMap 获取单一交易对清算地图
// 💥 对应前台「清算地图」页（/pro/futures/LiquidationMap）
// 🔒 需要登录：调用前须先 SetLoginToken，否则返回 ErrLoginRequired
// 🔐 自动生成 data 签名参数（TOTP + AES，见 generateLiqSign）
//
// 参数:
//
//	symbol    - 🪙 交易对，格式为「交易所_交易对」，如 Binance_BTCUSDT
//	rangeDays - 📅 时间范围（天），须为 LiqMapRanges 的 key: 1 / 7 / 30 / 90 / 180 / 365
//
// 返回:
//
//	*LiquidationMap - 📊 清算地图，Levels 按价格区间升序
//	error           - ❌ 参数错误、需要登录、请求或解密失败
//
// 使用示例:
//
//	spider.SetLoginToken(os.Getenv("COINGLASS_OBE"))
//	m, err := spider.GetLiquidationMap("Binance_BTCUSDT", 1)
//	for _, l := range m.Levels { ... }  // l.Bucket / l.AmountUSD / l.Leverage
func (s *Spider) GetLiquidationMap(symbol string, rangeDays int) (*LiquidationMap, error) {
	sign, err := generateLiqSign(time.Now())
	if err != nil {
		return nil, fmt.Errorf("❌ 生成清算地图签名失败: %w", err)
	}

	raw, err := s.getLiqMap(LiqMapEndpoint, symbol, rangeDays, sign)
	if err != nil {
		return nil, err
	}
	return parseLiquidationMap(raw)
}

// parseLiquidationMap 将清算地图原始JSON解析为 LiquidationMap
// 📝 独立出来便于用保存的响应离线验证，不必重复请求
func parseLiquidationMap(raw []byte) (*LiquidationMap, error) {
	// 📦 原始结构: liqMapV2 为「价格区间 -> 档位数组」的map
	var resp struct {
		Instrument LiqInstrument         `json:"instrument"`
		LastPrice  float64               `json:"lastPrice"`
		LiqMapV2   map[string][]LiqLevel `json:"liqMapV2"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("❌ 解析清算地图失败: %w", err)
	}

	levels, err := flattenLiqMapV2(resp.LiqMapV2)
	if err != nil {
		return nil, err
	}

	return &LiquidationMap{
		Instrument: resp.Instrument,
		LastPrice:  resp.LastPrice,
		Levels:     levels,
	}, nil
}

// parseExchangeLiquidationMap 将交易所聚合清算地图原始JSON解析为 ExchangeLiquidationMap
// 📝 data 数组每个元素是一家交易所，结构与单一交易对的 liqMapV2 相同
func parseExchangeLiquidationMap(raw []byte) (*ExchangeLiquidationMap, error) {
	var resp struct {
		Instrument LiqInstrument `json:"instrument"`
		LastPrice  float64       `json:"lastPrice"`
		RangeHigh  float64       `json:"rangeHigh"`
		RangeLow   float64       `json:"rangeLow"`
		Data       []struct {
			Instrument LiqInstrument         `json:"instrument"`
			LiqMapV2   map[string][]LiqLevel `json:"liqMapV2"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("❌ 解析交易所清算地图失败: %w", err)
	}

	result := &ExchangeLiquidationMap{
		Instrument: resp.Instrument,
		LastPrice:  resp.LastPrice,
		RangeHigh:  resp.RangeHigh,
		RangeLow:   resp.RangeLow,
	}
	for _, d := range resp.Data {
		levels, err := flattenLiqMapV2(d.LiqMapV2)
		if err != nil {
			return nil, fmt.Errorf("❌ %s: %w", d.Instrument.ExName, err)
		}
		result.Exchanges = append(result.Exchanges, ExchangeLiqMap{
			Instrument: d.Instrument,
			Levels:     levels,
		})
	}

	return result, nil
}

// flattenLiqMapV2 展平「价格区间 -> 档位数组」的map
// 🪣 回填 Bucket，并按区间升序（同区间再按杠杆升序）排列
func flattenLiqMapV2(m map[string][]LiqLevel) ([]LiqLevel, error) {
	var result []LiqLevel
	for key, levels := range m {
		bucket, err := strconv.ParseFloat(key, 64)
		if err != nil {
			return nil, fmt.Errorf("❌ 无效的价格区间 %q: %w", key, err)
		}
		for _, l := range levels {
			l.Bucket = bucket
			result = append(result, l)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Bucket != result[j].Bucket {
			return result[i].Bucket < result[j].Bucket
		}
		return result[i].Leverage < result[j].Leverage
	})
	return result, nil
}

// GetExchangeLiquidationMap 获取交易所聚合清算地图
// 💥 对应清算地图页的「交易所」视图，按币种聚合各交易所
// 🔒 需要登录：调用前须先 SetLoginToken，否则返回 ErrLoginRequired
// 🏢 实测（2026-10-05）返回 Binance / OKX / Bybit 三家
// ⚠️ 聚合版每个档位只有价格与金额，Leverage 为 0、Tag 为空（服务端返回 null）
//
// 参数:
//
//	symbol    - 🪙 币种，如 BTC（不是交易对）
//	rangeDays - 📅 时间范围（天），同 GetLiquidationMap
//
// 返回:
//
//	*ExchangeLiquidationMap - 🏢 各交易所清算档位，Levels 按价格区间升序
//	error                   - ❌ 参数错误、需要登录、请求或解密失败
func (s *Spider) GetExchangeLiquidationMap(symbol string, rangeDays int) (*ExchangeLiquidationMap, error) {
	raw, err := s.getLiqMap(ExLiqMapEndpoint, symbol, rangeDays, "")
	if err != nil {
		return nil, err
	}
	return parseExchangeLiquidationMap(raw)
}

// getLiqMap 清算地图共用请求逻辑
// 📝 sign 为空时不带 data 参数（exLiqMap 不需要签名）
func (s *Spider) getLiqMap(endpoint, symbol string, rangeDays int, sign string) (json.RawMessage, error) {
	r, ok := LiqMapRanges[rangeDays]
	if !ok {
		return nil, fmt.Errorf("❌ 无效的时间范围: %d 天（可选: 1 / 7 / 30 / 90 / 180 / 365）", rangeDays)
	}
	if strings.TrimSpace(symbol) == "" {
		return nil, fmt.Errorf("❌ symbol 不能为空")
	}

	// 🚪 未登录直接拦截，不浪费一次请求
	if s.loginToken == "" {
		return nil, ErrLoginRequired
	}

	params := url.Values{}
	params.Set("merge", "true")
	params.Set("symbol", symbol)
	params.Set("interval", r.Interval)
	params.Set("limit", strconv.Itoa(r.Limit))
	if sign != "" {
		params.Set("data", sign)
	}

	data, err := s.GetData(BaseURL + endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("❌ 获取清算地图失败: %w", err)
	}

	if !json.Valid([]byte(data)) {
		return nil, fmt.Errorf("❌ 清算地图返回的不是合法JSON: %.100s", data)
	}
	return json.RawMessage(data), nil
}

// ResponseMeta 响应头中与解密相关的元数据
// 🔑 服务端通过这些字段告知客户端该用哪种密钥推导方式
type ResponseMeta struct {
	User string // 🔐 加密的动态密钥
	V    string // 🎛️ 密钥来源标识，决定 deriveKeySource 的分支
	Time string // ⏰ 服务端时间戳（v=2 时作为密钥来源）
}

// deriveKeySource 根据响应头 v 决定AES密钥的来源字符串
// 🎛️ 服务端会切换密钥推导方式，v 标识当前使用哪一种
//
// 对应前端 mn(response, urlPath) 的逻辑:
//
//	v=0  -> 请求时发送的 cache-ts-v2（毫秒时间戳）
//	v=1  -> API路径（从 /api 起、不含查询字符串）
//	v=2  -> 响应头 time
//	v=55 -> 固定值 170b070da9654622
//	v=66 -> 固定值 d6537d845a964081
//	v=77 -> 固定值 863f08689c97435b
//
// ⚠️ 历史变更：2026-09-23 服务端由 v=0 切换为 v=1，
// 若只实现 v=0 会解密失败并报「无效的PKCS7填充」
func (s *Spider) deriveKeySource(meta *ResponseMeta, apiURL, cacheTsV2 string) (string, error) {
	switch meta.V {
	case "0":
		return cacheTsV2, nil
	case "1":
		return extractAPIPath(apiURL), nil
	case "2":
		if meta.Time == "" {
			return "", fmt.Errorf("v=2 但响应缺少 time header")
		}
		return meta.Time, nil
	case "55":
		return KeySourceV55, nil
	case "66":
		return KeySourceV66, nil
	case "77":
		return KeySourceV77, nil
	default:
		return "", fmt.Errorf("❌ 未知的加密版本 v=%q，服务端可能又更新了推导方式（见 README 的说明）", meta.V)
	}
}

// extractAPIPath 从完整URL中截取用作密钥来源的路径片段
// 📐 取 /api 开始到查询字符串之前的部分，如
//
//	https://capi.coinglass.com/api/etf/flow?x=1 -> /api/etf/flow
//
// 📝 找不到 /api 时返回原始URL（与前端行为一致）
func extractAPIPath(rawURL string) string {
	idx := strings.Index(rawURL, "/api")
	if idx == -1 {
		return rawURL
	}
	path := rawURL[idx:]
	if q := strings.Index(path, "?"); q != -1 {
		path = path[:q]
	}
	return path
}
