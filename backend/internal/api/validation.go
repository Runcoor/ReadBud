package api

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// FieldError represents a single field validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// fieldLabels maps DTO JSON field names to Chinese display labels.
var fieldLabels = map[string]string{
	"username":          "用户名",
	"password":          "密码",
	"keyword":           "关键词",
	"audience":          "目标受众",
	"tone":              "写作风格",
	"target_words":      "目标字数",
	"image_mode":        "配图模式",
	"chart_mode":        "图表模式",
	"publish_mode":      "发布模式",
	"publish_at":        "发布时间",
	"wechat_account_id": "公众号账号",
	"provider_type":     "供应商类型",
	"provider_name":     "供应商名称",
	"config_json":       "配置信息",
	"secret_json":       "密钥信息",
	"name":              "名称",
	"app_id":            "AppID",
	"app_secret":        "AppSecret",
	"token_mode":        "令牌模式",
	"draft_id":          "草稿",
	"title":             "标题",
	"subtitle":          "副标题",
	"digest":            "摘要",
	"heading":           "标题",
	"text_md":           "正文",
	"html_fragment":     "HTML片段",
	"schedule_at":       "定时发布时间",
	"remark":            "备注",
}

// tagMessages maps validator tag names to Chinese message templates.
// {0} = field label, {1} = param value.
var tagMessages = map[string]string{
	"required": "{0}不能为空",
	"min":      "{0}长度不能少于{1}个字符",
	"max":      "{0}长度不能超过{1}个字符",
	"oneof":    "{0}的值不合法，允许值为: {1}",
	"email":    "{0}格式不正确",
	"url":      "{0}格式不正确",
	"len":      "{0}长度必须为{1}",
	"gt":       "{0}必须大于{1}",
	"gte":      "{0}必须大于或等于{1}",
	"lt":       "{0}必须小于{1}",
	"lte":      "{0}必须小于或等于{1}",
}

// HandleBindError parses a Gin binding error and sends a structured 422 response
// with per-field validation details in Chinese. Returns true if the error was handled.
func HandleBindError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	// Try to parse as validator.ValidationErrors
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// Not a validation error — likely malformed JSON
		BadRequest(c, "请求格式错误，请检查 JSON 格式")
		return true
	}

	fields := make([]FieldError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		fields = append(fields, FieldError{
			Field:   jsonFieldName(fe),
			Message: translateFieldError(fe),
		})
	}

	ValidationError(c, "输入参数验证失败", fields)
	return true
}

// jsonFieldName extracts the JSON tag name from a validator.FieldError.
func jsonFieldName(fe validator.FieldError) string {
	// fe.Field() returns the Go struct field name; convert to JSON snake_case
	name := fe.Field()
	// Common conversions
	replacer := strings.NewReplacer(
		"Username", "username",
		"Password", "password",
		"Keyword", "keyword",
		"Audience", "audience",
		"Tone", "tone",
		"TargetWords", "target_words",
		"ImageMode", "image_mode",
		"ChartMode", "chart_mode",
		"PublishMode", "publish_mode",
		"PublishAt", "publish_at",
		"WechatAccountID", "wechat_account_id",
		"ProviderType", "provider_type",
		"ProviderName", "provider_name",
		"ConfigJSON", "config_json",
		"SecretJSON", "secret_json",
		"Name", "name",
		"AppID", "app_id",
		"AppSecret", "app_secret",
		"TokenMode", "token_mode",
		"DraftID", "draft_id",
		"Title", "title",
		"Subtitle", "subtitle",
		"Digest", "digest",
		"Heading", "heading",
		"TextMD", "text_md",
		"HTMLFragment", "html_fragment",
		"ScheduleAt", "schedule_at",
		"IsDefault", "is_default",
		"Remark", "remark",
		"DraftPublicID", "draft_public_id",
	)
	return replacer.Replace(name)
}

// translateFieldError translates a single field error to a Chinese message.
func translateFieldError(fe validator.FieldError) string {
	jsonName := jsonFieldName(fe)
	label := jsonName
	if l, ok := fieldLabels[jsonName]; ok {
		label = l
	}

	tag := fe.Tag()
	param := fe.Param()

	// Handle special number-based validations for int fields
	if tag == "min" || tag == "max" {
		switch fe.Kind().String() {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64":
			if tag == "min" {
				return fmt.Sprintf("%s不能小于%s", label, param)
			}
			return fmt.Sprintf("%s不能大于%s", label, param)
		}
	}

	// Use template if available
	if tmpl, ok := tagMessages[tag]; ok {
		msg := strings.ReplaceAll(tmpl, "{0}", label)
		msg = strings.ReplaceAll(msg, "{1}", param)
		return msg
	}

	return fmt.Sprintf("%s验证失败 (%s)", label, tag)
}
