package valueobjects

import (
	"bytes"
	"encoding/binary"
	"net"
	"net/url"
	"strings"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/textutils"
)

// 头像策略常量（SEC-08）
const (
	// AvatarURLMaxLen 头像 URL 最大长度（按字符计；users.avatar 列宽 256）
	AvatarURLMaxLen = 256
	// AvatarImageMaxSide 头像图片最大边长（宽/高均不得超过，防解压炸弹面）
	AvatarImageMaxSide = 4096
)

// AvatarFormat 允许的头像图片格式
type AvatarFormat string

const (
	// AvatarFormatPNG PNG
	AvatarFormatPNG AvatarFormat = "png"
	// AvatarFormatJPEG JPEG（.jpg / .jpeg）
	AvatarFormatJPEG AvatarFormat = "jpeg"
	// AvatarFormatWebP WebP（VP8 / VP8L / VP8X）
	AvatarFormatWebP AvatarFormat = "webp"
)

// 对外文案（可回显，不含路径/内部原因）
var (
	errAvatarURLInvalid   = domerr.Publicf("头像链接不合法（仅支持 http/https 公网地址）")
	errAvatarFormat       = domerr.Publicf("头像格式不受支持（仅支持 PNG/JPEG/WebP）")
	errAvatarExtMismatch  = domerr.Publicf("头像文件内容与扩展名不一致")
	errAvatarDimensionBad = domerr.Publicf(
		"图片尺寸过大（最大 %d×%d）", AvatarImageMaxSide, AvatarImageMaxSide,
	)
)

// avatarExtFormats 扩展名 → 期望内容格式（白名单；SVG/GIF 等不在其中）
var avatarExtFormats = map[string]AvatarFormat{
	".png":  AvatarFormatPNG,
	".jpg":  AvatarFormatJPEG,
	".jpeg": AvatarFormatJPEG,
	".webp": AvatarFormatWebP,
}

// 图片魔数
var (
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	webpRiff  = []byte("RIFF")
	webpTag   = []byte("WEBP")
)

// AvatarFormatOfExt 返回扩展名对应的允许格式
// @param ext 扩展名（需已小写，含点）
// @return 格式 / 是否在白名单
func AvatarFormatOfExt(ext string) (AvatarFormat, bool) {
	format, ok := avatarExtFormats[ext]
	return format, ok
}

// ValidateAvatarURL 校验 URL 方式头像（SEC-08）。
// 服务端**不抓取**该 URL（只存字符串、由客户端加载），因此这里只做无出站成本的字符串约束：
//   - 仅 http/https；拒绝 javascript:/data:/file:、协议相对、无 scheme、带 userinfo
//   - 长度 ≤ AvatarURLMaxLen 字符、不含空白与控制字符
//   - host 为 IP 字面量时拒绝私网/环回/链路本地/元数据/组播/未指定地址（**不做 DNS 解析**）
//
// @param raw 原始 URL（调用方应先 TrimSpace 再保存）
// @return 合法返回 nil，否则返回可回显的业务错误
func ValidateAvatarURL(raw string) error {
	if raw == "" || textutils.RuneLength(raw) > AvatarURLMaxLen {
		return errAvatarURLInvalid
	}
	for _, r := range raw {
		if r <= 0x20 || r == 0x7F {
			return errAvatarURLInvalid
		}
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return errAvatarURLInvalid
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return errAvatarURLInvalid
	}
	if parsed.User != nil {
		return errAvatarURLInvalid
	}
	host := parsed.Hostname()
	if host == "" {
		return errAvatarURLInvalid
	}
	if strings.EqualFold(host, "localhost") {
		return errAvatarURLInvalid
	}
	// 仅对 IP 字面量做网段判断：不做 DNS 解析（服务端没有出站请求，解析反而新增出网面）
	if ip := net.ParseIP(host); ip != nil && isForbiddenAvatarHost(ip) {
		return errAvatarURLInvalid
	}
	return nil
}

// isForbiddenAvatarHost 判断 IP 是否属于禁止的头像主机网段
// @param ip 已解析的 IP 字面量
// @return 属于私网/环回/链路本地/元数据/组播/未指定 返回 true
func isForbiddenAvatarHost(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return true
	}
	// 运营商级 NAT 100.64.0.0/10
	if v4 := ip.To4(); v4 != nil && v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
		return true
	}
	return false
}

// ValidateAvatarImage 按内容校验头像字节（SEC-08）：
// 魔数判定格式（PNG/JPEG/WebP）、扩展名与内容一致、宽高均 ≤ AvatarImageMaxSide。
// 只解析头部取宽高，不做整图解码（天然规避解压炸弹）。
// @param ext 扩展名（已小写，含点）
// @param data 文件内容
// @return 识别出的格式；不合法返回可回显的业务错误
func ValidateAvatarImage(ext string, data []byte) (AvatarFormat, error) {
	expected, ok := AvatarFormatOfExt(ext)
	if !ok {
		return "", errAvatarFormat
	}
	format, width, height, err := sniffAvatarImage(data)
	if err != nil {
		return "", err
	}
	if format != expected {
		return "", errAvatarExtMismatch
	}
	if width <= 0 || height <= 0 || width > AvatarImageMaxSide || height > AvatarImageMaxSide {
		return "", errAvatarDimensionBad
	}
	return format, nil
}

// sniffAvatarImage 只读头部判定图片格式与宽高
// @param data 文件内容（至少包含头部）
// @return 格式 / 宽 / 高 / 错误（无法确定 ⇒ 错误，调用方按 fail-closed 拒绝）
func sniffAvatarImage(data []byte) (AvatarFormat, int, int, error) {
	switch {
	case bytes.HasPrefix(data, pngMagic):
		return sniffPNG(data)
	case bytes.HasPrefix(data, jpegMagic):
		return sniffJPEG(data)
	case len(data) >= 12 && bytes.Equal(data[0:4], webpRiff) && bytes.Equal(data[8:12], webpTag):
		return sniffWebP(data)
	default:
		return "", 0, 0, errAvatarFormat
	}
}

// sniffPNG 读取 IHDR 的宽高
// @param data PNG 字节
// @return 格式 / 宽 / 高 / 错误
func sniffPNG(data []byte) (AvatarFormat, int, int, error) {
	const ihdrOffset = 16 // 8 签名 + 4 长度 + 4 "IHDR"
	if len(data) < ihdrOffset+8 || !bytes.Equal(data[12:16], []byte("IHDR")) {
		return "", 0, 0, errAvatarFormat
	}
	width := int(binary.BigEndian.Uint32(data[ihdrOffset : ihdrOffset+4]))
	height := int(binary.BigEndian.Uint32(data[ihdrOffset+4 : ihdrOffset+8]))
	return AvatarFormatPNG, width, height, nil
}

// sniffJPEG 扫描到 SOF 段读取宽高（不解码像素数据）
// @param data JPEG 字节
// @return 格式 / 宽 / 高 / 错误
func sniffJPEG(data []byte) (AvatarFormat, int, int, error) {
	// 起始 SOI 已校验；从 i=2 起按段遍历
	i := 2
	for i+1 < len(data) {
		if data[i] != 0xFF {
			return "", 0, 0, errAvatarFormat
		}
		// 跳过填充的 0xFF
		for i < len(data) && data[i] == 0xFF {
			i++
		}
		if i >= len(data) {
			return "", 0, 0, errAvatarFormat
		}
		marker := data[i]
		i++
		// 无长度字段的标记
		if marker == 0x01 || (marker >= 0xD0 && marker <= 0xD8) {
			continue
		}
		if marker == 0xD9 || i+2 > len(data) {
			return "", 0, 0, errAvatarFormat
		}
		segLen := int(binary.BigEndian.Uint16(data[i : i+2]))
		if segLen < 2 {
			return "", 0, 0, errAvatarFormat
		}
		// SOF0..SOF15（排除 DHT/JPG/DAC：0xC4/0xC8/0xCC）
		if marker >= 0xC0 && marker <= 0xCF &&
			marker != 0xC4 && marker != 0xC8 && marker != 0xCC {
			if i+7 > len(data) {
				return "", 0, 0, errAvatarFormat
			}
			height := int(binary.BigEndian.Uint16(data[i+3 : i+5]))
			width := int(binary.BigEndian.Uint16(data[i+5 : i+7]))
			return AvatarFormatJPEG, width, height, nil
		}
		i += segLen
	}
	return "", 0, 0, errAvatarFormat
}

// sniffWebP 解析 VP8 / VP8L / VP8X 三种变体的宽高；其他变体一律拒绝（fail-closed）
// @param data WebP 字节
// @return 格式 / 宽 / 高 / 错误
func sniffWebP(data []byte) (AvatarFormat, int, int, error) {
	const chunkOffset = 12 // "RIFF"+size+"WEBP"
	if len(data) < chunkOffset+8 {
		return "", 0, 0, errAvatarFormat
	}
	chunk := string(data[chunkOffset : chunkOffset+4])
	switch chunk {
	case "VP8X":
		return sniffWebPVP8X(data)
	case "VP8 ":
		return sniffWebPVP8(data)
	case "VP8L":
		return sniffWebPVP8L(data)
	default:
		// ALPH/ANIM/ANMF/未知变体：无法在此路径可靠取画布尺寸 ⇒ 拒绝
		return "", 0, 0, errAvatarFormat
	}
}

// sniffWebPVP8X 解析扩展格式（画布宽高以 24 位小端存储，值为实际值-1）
func sniffWebPVP8X(data []byte) (AvatarFormat, int, int, error) {
	const (
		chunkSizeOffset = 16
		canvasOffset    = 24
	)
	if len(data) < canvasOffset+6 {
		return "", 0, 0, errAvatarFormat
	}
	chunkSize := int(binary.LittleEndian.Uint32(data[chunkSizeOffset : chunkSizeOffset+4]))
	if chunkSize < 10 {
		return "", 0, 0, errAvatarFormat
	}
	width := int(le24(data[canvasOffset : canvasOffset+3]))
	height := int(le24(data[canvasOffset+3 : canvasOffset+6]))
	return AvatarFormatWebP, width + 1, height + 1, nil
}

// le24 读取 24 位小端无符号整数
// @param b 3 字节
// @return 数值
func le24(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

// sniffWebPVP8 解析有损变体（帧头中的 14 位宽高）
func sniffWebPVP8(data []byte) (AvatarFormat, int, int, error) {
	const (
		chunkSizeOffset = 16
		startCodeOffset = 23 // 3 字节帧标记之后
		dimsOffset      = 26
	)
	if len(data) < dimsOffset+4 {
		return "", 0, 0, errAvatarFormat
	}
	chunkSize := int(binary.LittleEndian.Uint32(data[chunkSizeOffset : chunkSizeOffset+4]))
	if chunkSize < 10 {
		return "", 0, 0, errAvatarFormat
	}
	if data[startCodeOffset] != 0x9D || data[startCodeOffset+1] != 0x01 ||
		data[startCodeOffset+2] != 0x2A {
		return "", 0, 0, errAvatarFormat
	}
	width := int(binary.LittleEndian.Uint16(data[dimsOffset:dimsOffset+2]) & 0x3FFF)
	height := int(binary.LittleEndian.Uint16(data[dimsOffset+2:dimsOffset+4]) & 0x3FFF)
	return AvatarFormatWebP, width, height, nil
}

// sniffWebPVP8L 解析无损变体（14 位宽高，值为实际值-1）
func sniffWebPVP8L(data []byte) (AvatarFormat, int, int, error) {
	const (
		chunkSizeOffset = 16
		signatureOffset = 20
		bitsOffset      = 21
	)
	if len(data) < bitsOffset+4 {
		return "", 0, 0, errAvatarFormat
	}
	chunkSize := int(binary.LittleEndian.Uint32(data[chunkSizeOffset : chunkSizeOffset+4]))
	if chunkSize < 5 || data[signatureOffset] != 0x2F {
		return "", 0, 0, errAvatarFormat
	}
	bits := binary.LittleEndian.Uint32(data[bitsOffset : bitsOffset+4])
	width := int(bits&0x3FFF) + 1
	height := int((bits>>14)&0x3FFF) + 1
	return AvatarFormatWebP, width, height, nil
}
