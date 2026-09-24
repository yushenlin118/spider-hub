package coinglass

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
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
}

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
	if meta.User == "" {
		return nil, nil, fmt.Errorf("未找到user header")
	}

	// 📦 解析JSON响应体
	var response CoinglassResponse
	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return nil, nil, fmt.Errorf("解析响应失败: %v", err)
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
