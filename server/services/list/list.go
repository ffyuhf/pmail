package list

import (
	"fmt"
	"strings"

	"github.com/ffyuhf/pmail/consts"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/array"
	"github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"

	. "xorm.io/builder"
)

func GetEmailList(ctx *context.Context, tagInfo dto.SearchTag, keyword string, pop3List bool, offset, limit int) (emailList []*response.EmailResponseData, total int64) {
	return getList(ctx, tagInfo, keyword, pop3List, offset, limit)
}

func getList(ctx *context.Context, tagInfo dto.SearchTag, keyword string, pop3List bool, offset, limit int) (emailList []*response.EmailResponseData, total int64) {
	querySQL, queryParams := genSQL(ctx, false, tagInfo, keyword, pop3List, offset, limit)

	err := db.Instance.SQL(querySQL, queryParams...).Find(&emailList)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	totalSQL, totalParams := genSQL(ctx, true, tagInfo, keyword, pop3List, offset, limit)

	_, err = db.Instance.SQL(totalSQL, totalParams...).Get(&total)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL ERROR: %s ,Error:%s", querySQL, err)
	}

	return emailList, total
}

func genSQL(ctx *context.Context, count bool, tagInfo dto.SearchTag, keyword string, pop3List bool, offset, limit int) (string, []any) {
	sqlParams := []any{ctx.UserID}
	sql := "select "

	if count {
		sql += `count(1) from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? `
	} else if pop3List {
		sql += `e.id,e.size from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? `
	} else {
		sql += `e.*,ue.is_read,ue.id as ue_id from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? `
	}

	// 移植日期: 20260815 — 系统文件夹（已删除/草稿/垃圾）查询改为 status 或 group_id 双条件命中，
	// 与 doMove/Move2DefaultBox 写入系统文件夹语义配套（母项目改进合并移植，保留现有其余行为）
	if tagInfo.Status != -1 {
		switch tagInfo.Status {
		case consts.EmailStatusDel:
			sql += " and (ue.status =? or ue.group_id=?) "
			sqlParams = append(sqlParams, tagInfo.Status, models.Deleted)
		case consts.EmailStatusDrafts:
			sql += " and (ue.status =? or ue.group_id=?) "
			sqlParams = append(sqlParams, tagInfo.Status, models.Drafts)
		case consts.EmailStatusJunk:
			sql += " and (ue.status =? or ue.group_id=?) "
			sqlParams = append(sqlParams, tagInfo.Status, models.Junk)
		default:
			sql += " and ue.status =? "
			sqlParams = append(sqlParams, tagInfo.Status)
		}
	} else if tagInfo.Status == -1 {
		// 修复日期: 20260516 — 条件从 != -1 改为 > 0，避免 GroupId=0 的内置视图（收件箱/发件箱）跳过 status 过滤
		// 根因：收件箱 Tag 的 GroupId=0，原条件 0!=-1 为 true 导致跳过 status 过滤，软删除后邮件仍显示在原文件夹
		if tagInfo.GroupId > 0 {
			// 自定义分组视图（GroupId > 0）：不限制 status，展示该分组下所有状态的邮件
		} else if tagInfo.Type != 1 {
			sql += " and ue.status = 0"
		} else {
			// 发件箱不展示已删除的邮件
			sql += " and ue.status != 3"
		}
	}

	// 移植日期: 20260815 — 发件箱查询改为 type 或 Sent 系统文件夹 group_id 双条件命中（母项目改进合并移植）
	if tagInfo.Type != -1 {
		if tagInfo.Type == consts.EmailTypeSend {
			sql += " and (type =? or ue.group_id=?)"
			sqlParams = append(sqlParams, tagInfo.Type, models.Sent)
		} else {
			sql += " and type =? "
			sqlParams = append(sqlParams, tagInfo.Type)
		}
	}

	// 移植日期: 20260815 — GroupId 过滤适配系统文件夹：收件箱视图命中 group_id=0 或 INBOX，
	// 全部邮件视图（GroupId 不限）命中默认文件夹或 INBOX（母项目改进合并移植）
	if tagInfo.GroupId != -1 {
		if tagInfo.GroupId > 0 {
			sql += " and ue.group_id=? "
			sqlParams = append(sqlParams, tagInfo.GroupId)
		} else if tagInfo.GroupId == 0 && tagInfo.Status == -1 {
			sql += " and (ue.group_id=? or ue.group_id=0)"
			sqlParams = append(sqlParams, models.INBOX)
		}
	} else {
		sql += " and (ue.group_id=0 or ue.group_id =?) "
		sqlParams = append(sqlParams, models.INBOX)
	}

	// 修改日期: 20260610 — #15 修复搜索关键词 SQL LIKE 通配符注入
	// 转义 % 和 _ 防止用户输入的通配符干扰查询语义
	if keyword != "" {
		escapedKeyword := strings.ReplaceAll(keyword, "%", `\%`)
		escapedKeyword = strings.ReplaceAll(escapedKeyword, "_", `\_`)
		sql += " and (subject like ? or text like ? )"
		sqlParams = append(sqlParams, "%"+escapedKeyword+"%", "%"+escapedKeyword+"%")
	}

	if limit == 0 {
		limit = 10
	}

	sql += " order by e.id desc"

	if limit < 10000 {
		sql += fmt.Sprintf(" LIMIT %d OFFSET %d ", limit, offset)
	}

	return sql, sqlParams

}

type statRes struct {
	Total int64
	Size  int64
}

// Stat 查询邮件总数和大小
func Stat(ctx *context.Context) (int64, int64) {
	sql := `select count(1) as total,sum(size) as size from email e left join user_email ue on e.id=ue.email_id where ue.user_id = ? and e.type = 0 and ue.status != 3`
	var ret statRes
	_, err := db.Instance.SQL(sql, ctx.UserID).Get(&ret)
	if err != nil {
		log.WithContext(ctx).Errorf("SQL ERROR: %s ,Error:%s", sql, err)
	}
	return ret.Total, ret.Size
}

type ImapListReq struct {
	UidList []int
	Star    int
	End     int
}

func GetUEListByUID(ctx *context.Context, groupName string, star, end int, uidList []int) []*response.UserEmailUIDData {
	var ue []*response.UserEmailUIDData
	sql := "SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE user_id = ? "

	params := []any{ctx.UserID}

	if len(uidList) > 0 {
		sql += fmt.Sprintf(" and id in (%s)", array.Join(uidList, ","))
	}
	if star > 0 {
		sql += " and id >=?"
		params = append(params, star)
	}
	if end > 0 {
		sql += " and id <=?"
		params = append(params, end)
	}

	switch groupName {
	case "INBOX":
		sql += " and status =?"
		params = append(params, 0)
	case "Sent Messages":
		sql += " and status =?"
		params = append(params, 1)
	case "Drafts":
		sql += " and status =?"
		params = append(params, 4)
	case "Deleted Messages":
		sql += " and status =?"
		params = append(params, 3)
	case "Junk":
		sql += " and status =?"
		params = append(params, 5)
	default:
		groupNames := strings.Split(groupName, "/")
		groupName = groupNames[len(groupNames)-1]

		var group models.Group
		db.Instance.Table("group").Where("user_id=? and name=?", ctx.UserID, groupName).Get(&group)
		if group.ID == 0 {
			return nil
		}

		sql += " and group_id = ?"
		params = append(params, group.ID)
	}

	db.Instance.SQL(sql, params...).Find(&ue)
	return ue
}

func getEmailListByUidList(ctx *context.Context, groupName string, req ImapListReq, uid bool) []*response.EmailResponseData {
	var ret []*response.EmailResponseData
	var ue []*response.UserEmailUIDData
	sql := fmt.Sprintf("SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and id in (%s) and status = ?)", array.Join(req.UidList, ","))
	if req.Star > 0 && req.End != 0 {
		sql = fmt.Sprintf("SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and id >=%d and id <= %d and status = ?)", req.Star, req.End)
	}
	if req.Star > 0 && req.End == 0 {
		sql = fmt.Sprintf("SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and id >=%d and status = ?)", req.Star)
	}

	var err error
	switch groupName {
	case "INBOX":
		err = db.Instance.SQL(sql, ctx.UserID, 0).Find(&ue)
	case "Sent Messages":
		err = db.Instance.SQL(sql, ctx.UserID, 1).Find(&ue)
	case "Drafts":
		err = db.Instance.SQL(sql, ctx.UserID, 4).Find(&ue)
	case "Deleted Messages":
		err = db.Instance.SQL(sql, ctx.UserID, 3).Find(&ue)
	case "Junk":
		err = db.Instance.SQL(sql, ctx.UserID, 5).Find(&ue)
	default:
		groupNames := strings.Split(groupName, "/")
		groupName = groupNames[len(groupNames)-1]

		var group models.Group
		db.Instance.Table("group").Where("user_id=? and name=?", ctx.UserID, groupName).Get(&group)
		if group.ID == 0 {
			return ret
		}
		err = db.Instance.
			SQL(fmt.Sprintf(
				"SELECT * from (SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and group_id = ?)) a WHERE serial_number in (%s)",
				array.Join(req.UidList, ","))).
			Find(&ue, ctx.UserID, group.ID)
	}

	if err != nil {
		log.WithContext(ctx).Errorf("SQL ERROR: %s ,Error:%s", sql, err)
	}
	ueMap := map[int]*response.UserEmailUIDData{}
	var emailIds []int
	for _, email := range ue {
		ueMap[email.EmailID] = email
		emailIds = append(emailIds, email.EmailID)
	}

	_ = db.Instance.Table("email").Select("*").Where(Eq{"id": emailIds}).Find(&ret)
	for i, data := range ret {
		ret[i].IsRead = ueMap[data.Id].IsRead
		ret[i].SerialNumber = ueMap[data.Id].SerialNumber
		ret[i].UeId = ueMap[data.Id].ID
	}

	return ret
}

func GetEmailListByGroup(ctx *context.Context, groupName string, req ImapListReq, uid bool) []*response.EmailResponseData {
	if len(req.UidList) == 0 && req.Star == 0 && req.End == 0 {
		return nil
	}

	if uid {
		return getEmailListByUidList(ctx, groupName, req, uid)
	}

	var ret []*response.EmailResponseData
	var ue []*response.UserEmailUIDData

	sql := fmt.Sprintf("SELECT * from (SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and status = ? and group_id=0 )) a WHERE serial_number in (%s)", array.Join(req.UidList, ","))
	if req.Star > 0 && req.End == 0 {
		sql = fmt.Sprintf("SELECT * from (SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and status = ? and group_id=0 )) a WHERE serial_number >= %d", req.Star)
	}
	if req.Star > 0 && req.End > 0 {
		sql = fmt.Sprintf("SELECT * from (SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and status = ? and group_id=0 )) a WHERE serial_number >= %d and serial_number <=%d", req.Star, req.End)
	}

	switch groupName {
	case "INBOX":
		db.Instance.SQL(sql, ctx.UserID, 0).Find(&ue)
	case "Sent Messages":
		db.Instance.SQL(sql, ctx.UserID, 1).Find(&ue)
	case "Drafts":
		db.Instance.SQL(sql, ctx.UserID, 4).Find(&ue)
	case "Deleted Messages":
		db.Instance.SQL(sql, ctx.UserID, 3).Find(&ue)
	case "Junk":
		db.Instance.SQL(sql, ctx.UserID, 5).Find(&ue)
	default:
		groupNames := strings.Split(groupName, "/")
		groupName = groupNames[len(groupNames)-1]

		var group models.Group
		db.Instance.Table("group").Where("user_id=? and name=?", ctx.UserID, groupName).Get(&group)
		if group.ID == 0 {
			return ret
		}
		db.Instance.
			SQL(fmt.Sprintf(
				"SELECT * from (SELECT id,email_id, is_read, ROW_NUMBER() OVER (ORDER BY id) AS serial_number FROM `user_email` WHERE (user_id = ? and group_id = ?)) a WHERE serial_number in (%s)",
				array.Join(req.UidList, ","))).
			Find(&ue, ctx.UserID, group.ID)
	}

	ueMap := map[int]*response.UserEmailUIDData{}
	var emailIds []int
	for _, email := range ue {
		ueMap[email.EmailID] = email
		emailIds = append(emailIds, email.EmailID)
	}

	_ = db.Instance.Table("email").Select("*").Where(Eq{"id": emailIds}).Find(&ret)
	for i, data := range ret {
		ret[i].IsRead = ueMap[data.Id].IsRead
		ret[i].SerialNumber = ueMap[data.Id].SerialNumber
		ret[i].UeId = ueMap[data.Id].ID
	}

	return ret
}
