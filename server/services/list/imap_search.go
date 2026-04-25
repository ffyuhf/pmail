package list

import (
	"strings"
	"time"

	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/emersion/go-imap/v2"
	log "github.com/sirupsen/logrus"
)

// IMAPSearchCriteria 表示 IMAP SEARCH 命令的搜索条件
type IMAPSearchCriteria struct {
	// UID 和序列号集合
	UIDs    []imap.UIDSet
	SeqNums []imap.SeqSet

	// 日期过滤（仅使用日期，时间和时区被忽略）
	Since      time.Time // 内部日期起始
	Before     time.Time // 内部日期截止
	SentSince  time.Time // 发送日期起始
	SentBefore time.Time // 发送日期截止

	// 头部字段搜索
	Header []HeaderField // 用于头部搜索的键值对

	// 内容搜索
	Body []string // 在正文中搜索
	Text []string // 在头部和正文中搜索

	// 标志过滤
	Flag    []imap.Flag // 包含这些标志的邮件
	NotFlag []imap.Flag // 不包含这些标志的邮件

	// 大小过滤
	Larger  int64 // 大于此大小的邮件（字节）
	Smaller int64 // 小于此大小的邮件（字节）

	// 逻辑组合
	Not []IMAPSearchCriteria
	Or  [][2]IMAPSearchCriteria
}

// HeaderField 表示头部字段搜索条件
type HeaderField struct {
	Key   string
	Value string
}

// IMAPSearchResult 表示搜索结果项
type IMAPSearchResult struct {
	UID          int  // user_email.id 作为 UID
	SeqNum       int  // 序列号
	EmailID      int  // email.id
	IsRead       int8 // 已读状态
	SerialNumber int  // 序号
}

// SearchEmails 根据条件执行 IMAP 搜索
func SearchEmails(ctx *context.Context, groupName string, criteria *imap.SearchCriteria) ([]*response.UserEmailUIDData, error) {
	// 首先获取邮箱的基础列表
	baseList := GetUEListByUID(ctx, groupName, 0, 0, nil)
	if len(baseList) == 0 {
		return baseList, nil
	}

	// 如果没有指定搜索条件，返回全部
	if criteria == nil || isEmptyCriteria(criteria) {
		return baseList, nil
	}

	// 构建 UID 到序列号的映射
	uidToSeq := make(map[int]int)
	seqToUID := make(map[int]int)
	for _, item := range baseList {
		uidToSeq[item.ID] = item.SerialNumber
		seqToUID[item.SerialNumber] = item.ID
	}

	// 按 UID 集合过滤
	if len(criteria.UID) > 0 {
		baseList = filterByUIDSets(baseList, criteria.UID)
	}

	// 按序列号集合过滤
	if len(criteria.SeqNum) > 0 {
		baseList = filterBySeqNumSets(baseList, criteria.SeqNum)
	}

	// 对于更复杂的过滤，需要获取邮件数据
	if needsEmailData(criteria) {
		baseList = filterWithEmailData(ctx, baseList, criteria)
	}

	// 按标志过滤（本实现中为已读状态）
	if len(criteria.Flag) > 0 || len(criteria.NotFlag) > 0 {
		baseList = filterByFlags(baseList, criteria.Flag, criteria.NotFlag)
	}

	// 处理 NOT 条件
	if len(criteria.Not) > 0 {
		for _, notCriteria := range criteria.Not {
			baseList = applyNotCriteria(ctx, groupName, baseList, &notCriteria)
		}
	}

	// 处理 OR 条件
	if len(criteria.Or) > 0 {
		baseList = applyOrCriteria(ctx, groupName, baseList, criteria.Or)
	}

	return baseList, nil
}

// isEmptyCriteria 检查搜索条件是否为空
func isEmptyCriteria(criteria *imap.SearchCriteria) bool {
	return len(criteria.UID) == 0 &&
		len(criteria.SeqNum) == 0 &&
		criteria.Since.IsZero() &&
		criteria.Before.IsZero() &&
		criteria.SentSince.IsZero() &&
		criteria.SentBefore.IsZero() &&
		len(criteria.Header) == 0 &&
		len(criteria.Body) == 0 &&
		len(criteria.Text) == 0 &&
		len(criteria.Flag) == 0 &&
		len(criteria.NotFlag) == 0 &&
		criteria.Larger == 0 &&
		criteria.Smaller == 0 &&
		len(criteria.Not) == 0 &&
		len(criteria.Or) == 0
}

// needsEmailData 检查是否需要加载邮件数据进行过滤
func needsEmailData(criteria *imap.SearchCriteria) bool {
	return !criteria.Since.IsZero() ||
		!criteria.Before.IsZero() ||
		!criteria.SentSince.IsZero() ||
		!criteria.SentBefore.IsZero() ||
		len(criteria.Header) > 0 ||
		len(criteria.Body) > 0 ||
		len(criteria.Text) > 0 ||
		criteria.Larger > 0 ||
		criteria.Smaller > 0
}

// filterByUIDSets 按 UID 集合过滤列表
func filterByUIDSets(list []*response.UserEmailUIDData, uidSets []imap.UIDSet) []*response.UserEmailUIDData {
	var result []*response.UserEmailUIDData
	for _, item := range list {
		for _, uidSet := range uidSets {
			if uidSet.Contains(imap.UID(item.ID)) {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

// filterBySeqNumSets 按序列号集合过滤列表
func filterBySeqNumSets(list []*response.UserEmailUIDData, seqSets []imap.SeqSet) []*response.UserEmailUIDData {
	var result []*response.UserEmailUIDData
	for _, item := range list {
		for _, seqSet := range seqSets {
			if seqSet.Contains(uint32(item.SerialNumber)) {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

// filterByFlags 按邮件标志过滤
func filterByFlags(list []*response.UserEmailUIDData, flags []imap.Flag, notFlags []imap.Flag) []*response.UserEmailUIDData {
	var result []*response.UserEmailUIDData
	for _, item := range list {
		match := true

		// 检查必须存在的标志
		for _, flag := range flags {
			if !hasFlag(item, flag) {
				match = false
				break
			}
		}

		// 检查不应存在的标志
		if match {
			for _, flag := range notFlags {
				if hasFlag(item, flag) {
					match = false
					break
				}
			}
		}

		if match {
			result = append(result, item)
		}
	}
	return result
}

// hasFlag 检查邮件是否具有特定标志
func hasFlag(item *response.UserEmailUIDData, flag imap.Flag) bool {
	switch flag {
	case imap.FlagSeen:
		return item.IsRead == 1
	case imap.FlagDeleted:
		return item.Status == 3
	case imap.FlagDraft:
		return item.Status == 4
	case imap.FlagJunk:
		return item.Status == 5
	// 对于未跟踪的标志，返回 false
	case imap.FlagAnswered, imap.FlagFlagged:
		return false
	default:
		return false
	}
}

// filterWithEmailData 加载邮件数据并应用需要邮件数据的过滤器
func filterWithEmailData(ctx *context.Context, list []*response.UserEmailUIDData, criteria *imap.SearchCriteria) []*response.UserEmailUIDData {
	if len(list) == 0 {
		return list
	}

	// 获取邮件 ID
	var emailIDs []int
	ueMap := make(map[int]*response.UserEmailUIDData) // emailID -> UserEmailUIDData
	for _, item := range list {
		emailIDs = append(emailIDs, item.EmailID)
		ueMap[item.EmailID] = item
	}

	// 从数据库获取邮件
	var emails []models.Email
	err := db.Instance.Table("email").In("id", emailIDs).Find(&emails)
	if err != nil {
		log.WithContext(ctx).Errorf("Failed to fetch emails for search: %v", err)
		return list
	}

	// 构建邮件映射
	emailMap := make(map[int]*models.Email)
	for i := range emails {
		emailMap[emails[i].Id] = &emails[i]
	}

	// 过滤
	var result []*response.UserEmailUIDData
	for _, item := range list {
		email, ok := emailMap[item.EmailID]
		if !ok {
			continue
		}

		if matchesEmailCriteria(email, criteria) {
			result = append(result, item)
		}
	}

	return result
}

// matchesEmailCriteria 检查邮件是否匹配搜索条件
func matchesEmailCriteria(email *models.Email, criteria *imap.SearchCriteria) bool {
	// 日期过滤（内部日期 = CreateTime）
	if !criteria.Since.IsZero() {
		if email.CreateTime.Before(truncateToDate(criteria.Since)) {
			return false
		}
	}
	if !criteria.Before.IsZero() {
		if !email.CreateTime.Before(truncateToDate(criteria.Before)) {
			return false
		}
	}

	// 发送日期过滤
	if !criteria.SentSince.IsZero() {
		if email.SendDate.Before(truncateToDate(criteria.SentSince)) {
			return false
		}
	}
	if !criteria.SentBefore.IsZero() {
		if !email.SendDate.Before(truncateToDate(criteria.SentBefore)) {
			return false
		}
	}

	// 大小过滤
	if criteria.Larger > 0 {
		if int64(email.Size) <= criteria.Larger {
			return false
		}
	}
	if criteria.Smaller > 0 {
		if int64(email.Size) >= criteria.Smaller {
			return false
		}
	}

	// 头部字段搜索
	for _, hf := range criteria.Header {
		if !matchesHeader(email, hf.Key, hf.Value) {
			return false
		}
	}

	// 正文搜索
	for _, pattern := range criteria.Body {
		if !matchesBody(email, pattern) {
			return false
		}
	}

	// 全文搜索（头部 + 正文）
	for _, pattern := range criteria.Text {
		if !matchesText(email, pattern) {
			return false
		}
	}

	return true
}

// truncateToDate 去除 time.Time 的时间部分，仅保留日期
func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// matchesHeader 检查邮件是否匹配头部字段搜索
func matchesHeader(email *models.Email, key, value string) bool {
	key = strings.ToLower(key)
	value = strings.ToLower(value)

	switch key {
	case "subject":
		return strings.Contains(strings.ToLower(email.Subject), value)
	case "from":
		return strings.Contains(strings.ToLower(email.FromAddress), value) ||
			strings.Contains(strings.ToLower(email.FromName), value)
	case "to":
		return strings.Contains(strings.ToLower(email.To), value)
	case "cc":
		return strings.Contains(strings.ToLower(email.Cc), value)
	case "bcc":
		return strings.Contains(strings.ToLower(email.Bcc), value)
	case "reply-to":
		return strings.Contains(strings.ToLower(email.ReplyTo), value)
	case "sender":
		return strings.Contains(strings.ToLower(email.Sender), value)
	default:
		// 对于未知的头部字段，无法匹配
		return false
	}
}

// matchesBody 检查邮件正文是否匹配模式
func matchesBody(email *models.Email, pattern string) bool {
	pattern = strings.ToLower(pattern)

	// 检查文本正文
	if email.Text.Valid && strings.Contains(strings.ToLower(email.Text.String), pattern) {
		return true
	}

	// 检查 HTML 正文
	if email.Html.Valid && strings.Contains(strings.ToLower(email.Html.String), pattern) {
		return true
	}

	return false
}

// matchesText 检查邮件（头部 + 正文）是否匹配模式
func matchesText(email *models.Email, pattern string) bool {
	pattern = strings.ToLower(pattern)

	// 检查头部
	if strings.Contains(strings.ToLower(email.Subject), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.FromAddress), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.FromName), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.To), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.Cc), pattern) {
		return true
	}
	if strings.Contains(strings.ToLower(email.Bcc), pattern) {
		return true
	}

	// 检查正文
	return matchesBody(email, pattern)
}

// applyNotCriteria 应用 NOT 条件
func applyNotCriteria(ctx *context.Context, groupName string, list []*response.UserEmailUIDData, notCriteria *imap.SearchCriteria) []*response.UserEmailUIDData {
	// 获取匹配 NOT 条件的项目列表
	matchedList, _ := SearchEmails(ctx, groupName, notCriteria)

	// 构建已匹配的 UID 集合
	matchedUIDs := make(map[int]bool)
	for _, item := range matchedList {
		matchedUIDs[item.ID] = true
	}

	// 返回不在匹配集合中的项目
	var result []*response.UserEmailUIDData
	for _, item := range list {
		if !matchedUIDs[item.ID] {
			result = append(result, item)
		}
	}

	return result
}

// applyOrCriteria 应用 OR 条件
func applyOrCriteria(ctx *context.Context, groupName string, list []*response.UserEmailUIDData, orCriteria [][2]imap.SearchCriteria) []*response.UserEmailUIDData {
	if len(orCriteria) == 0 {
		return list
	}

	// 构建当前 UID 集合用于交集运算
	currentUIDs := make(map[int]bool)
	for _, item := range list {
		currentUIDs[item.ID] = true
	}

	// 对于每个 OR 对，查找匹配任一条件的项目
	resultUIDs := make(map[int]bool)

	for _, pair := range orCriteria {
		// 获取匹配第一个条件的结果
		matches1, _ := SearchEmails(ctx, groupName, &pair[0])
		for _, item := range matches1 {
			if currentUIDs[item.ID] {
				resultUIDs[item.ID] = true
			}
		}

		// 获取匹配第二个条件的结果
		matches2, _ := SearchEmails(ctx, groupName, &pair[1])
		for _, item := range matches2 {
			if currentUIDs[item.ID] {
				resultUIDs[item.ID] = true
			}
		}
	}

	// 从原始列表构建结果，保持顺序
	var result []*response.UserEmailUIDData
	for _, item := range list {
		if resultUIDs[item.ID] {
			result = append(result, item)
		}
	}

	return result
}
