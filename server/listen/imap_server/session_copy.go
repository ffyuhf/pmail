package imap_server

import (
	"github.com/emersion/go-imap/v2"
	"github.com/ffyuhf/pmail/consts"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/services/group"
	"github.com/ffyuhf/pmail/services/list"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/spf13/cast"
)

func (s *serverSession) Copy(numSet imap.NumSet, dest string) (*imap.CopyData, error) {

	var emailList []*response.EmailResponseData

	switch numSet.(type) {
	case imap.SeqSet:
		seqSet := numSet.(imap.SeqSet)
		for _, seq := range seqSet {
			emailList = list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(seq.Start),
				End:  cast.ToInt(seq.Stop),
			}, false)
		}
	case imap.UIDSet:
		uidSet := numSet.(imap.UIDSet)
		for _, uid := range uidSet {
			emailList = list.GetEmailListByGroup(s.ctx, s.currentMailbox, list.ImapListReq{
				Star: cast.ToInt(uint32(uid.Start)),
				End:  cast.ToInt(uint32(uid.Stop)),
			}, true)
		}
	}

	if len(emailList) == 0 {
		return nil, &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "Email Not Found",
		}
	}

	var err error
	destUid := []int{}
	UIDValidity := 0
	if group.IsDefaultBox(dest) {
		UIDValidity, destUid, err = copy2defaultbox(s.ctx, emailList, dest)
	} else {
		UIDValidity, destUid, err = copy2userbox(s.ctx, emailList, dest)
	}
	data := imap.CopyData{}
	data.UIDValidity = cast.ToUint32(UIDValidity)
	data.DestUIDs = imap.UIDSet{}
	data.SourceUIDs = imap.UIDSet{}
	for _, uid := range destUid {
		data.DestUIDs = append(data.DestUIDs, imap.UIDRange{Start: imap.UID(cast.ToUint32(uid)), Stop: imap.UID(cast.ToUint32(uid))})
	}

	for _, email := range emailList {
		data.SourceUIDs = append(data.SourceUIDs, imap.UIDRange{Start: imap.UID(cast.ToUint32(email.UeId)), Stop: imap.UID(cast.ToUint32(email.UeId))})
	}

	return &data, err
}

// copy2defaultbox 复制邮件到默认文件夹
// 修改日期: 20260510 — 添加重复检查，避免同一 email_id+status+group_id 重复创建记录
func copy2defaultbox(ctx *context.Context, mails []*response.EmailResponseData, dest string) (int, []int, error) {

	var destUid []int
	for _, email := range mails {
		var status int8
		switch dest {
		case "Deleted Messages":
			status = consts.EmailStatusDel
		case "INBOX":
			status = consts.EmailStatusWait
		case "Sent Messages":
			status = consts.EmailStatusSent
		case "Drafts":
			status = consts.EmailStatusDrafts
		case "Junk":
			status = consts.EmailStatusJunk
		}

		// 检查是否已存在相同的 email_id + status + group_id 记录
		var existCount int
		db.Instance.Table(&models.UserEmail{}).Select("count(1)").
			Where("user_id=? and email_id=? and status=? and group_id=0", ctx.UserID, email.Id, status).
			Get(&existCount)
		if existCount > 0 {
			continue // 跳过重复记录
		}

		newUe := models.UserEmail{
			UserID:  ctx.UserID,
			EmailID: email.Id,
			IsRead:  email.IsRead,
			GroupId: 0,
			Status:  status,
		}
		db.Instance.Insert(&newUe)
		destUid = append(destUid, newUe.ID)
	}

	return models.GroupNameToCode[dest], destUid, nil
}

// copy2userbox 复制邮件到用户自定义文件夹
// 修改日期: 20260510 — 添加重复检查，避免同一 email_id+group_id 重复创建记录
func copy2userbox(ctx *context.Context, mails []*response.EmailResponseData, dest string) (int, []int, error) {
	groupInfo, err := group.GetGroupByFullPath(ctx, dest)
	if err != nil {
		return 0, nil, &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: err.Error(),
		}
	}
	if groupInfo == nil || groupInfo.ID == 0 {
		return 0, nil, &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "Group not found",
		}
	}

	// 检查目标分组是否含有子文件夹，含有子文件夹的分组不允许放置邮件
	if group.HasChildren(ctx, groupInfo.ID) {
		return 0, nil, &imap.Error{
			Type: imap.StatusResponseTypeNo,
			Text: "Cannot copy mail to a folder that contains subfolders",
		}
	}

	var destUid []int
	for _, email := range mails {
		// 检查是否已存在相同的 email_id + group_id 记录
		var existCount int
		db.Instance.Table(&models.UserEmail{}).Select("count(1)").
			Where("user_id=? and email_id=? and group_id=?", ctx.UserID, email.Id, groupInfo.ID).
			Get(&existCount)
		if existCount > 0 {
			continue // 跳过重复记录
		}

		newUe := models.UserEmail{
			UserID:  ctx.UserID,
			EmailID: email.Id,
			IsRead:  email.IsRead,
			GroupId: groupInfo.ID,
			Status:  email.Status,
		}
		db.Instance.Insert(&newUe)
		destUid = append(destUid, newUe.ID)
	}

	return groupInfo.ID, destUid, nil
}
