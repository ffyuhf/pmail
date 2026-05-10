package del_email

import (
	"github.com/ffyuhf/pmail/consts"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"xorm.io/xorm"

	. "xorm.io/builder"
)

// DelEmailByUeIds 通过 user_email.id（ueIds）精确删除记录
// 修改日期: 20260510 — 新增函数，避免删除跨文件夹影响其他副本
func DelEmailByUeIds(ctx *context.Context, ueIds []int, forcedDel bool) error {
	session := db.Instance.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}
	for _, ueId := range ueIds {
		err := deleteOneByUeId(ctx, session, cast.ToInt64(ueId), forcedDel)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return session.Commit()
}

// deleteOneByUeId 通过 user_email.id 精确定位并删除一条记录
// 修改日期: 20260510 — 仅影响指定的 user_email 记录，不影响其他文件夹中的副本
func deleteOneByUeId(ctx *context.Context, session *xorm.Session, ueId int64, forcedDel bool) error {
	if !forcedDel {
		// 软删除：仅更新指定的 user_email 记录
		_, err := session.Table(&models.UserEmail{}).Where("id=? and user_id=?", ueId, ctx.UserID).Update(map[string]interface{}{
			"status":   consts.EmailStatusDel,
			"group_id": 0,
		})
		return err
	}
	// 硬删除：先获取 email_id，再删除 user_email 记录，最后检查是否还有其他关联
	var ue models.UserEmail
	_, err := session.Table(&models.UserEmail{}).Where("id=? and user_id=?", ueId, ctx.UserID).Get(&ue)
	if err != nil || ue.ID == 0 {
		return err
	}
	emailId := ue.EmailID

	_, err = session.Table(&models.UserEmail{}).Where("id=? and user_id=?", ueId, ctx.UserID).Delete(&ue)
	if err != nil {
		return err
	}
	// 检查 email 是否还有其他 user_email 关联
	var Num num
	_, err = session.Table(&models.UserEmail{}).Select("count(1) as num").Where("email_id=? ", emailId).Get(&Num)
	if err != nil {
		return err
	}
	if Num.Num == 0 {
		var email models.Email
		_, err = session.Table(&email).Where("id=?", emailId).Delete(&email)
	}
	return err
}

func DelEmail(ctx *context.Context, ids []int, forcedDel bool) error {
	session := db.Instance.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return err
	}
	for _, id := range ids {
		err := deleteOne(ctx, session, cast.ToInt64(id), forcedDel)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return session.Commit()
}

type num struct {
	Num int `xorm:"num"`
}

func deleteOne(ctx *context.Context, session *xorm.Session, id int64, forcedDel bool) error {
	if !forcedDel {
		_, err := session.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", id, ctx.UserID).Update(map[string]interface{}{
			"status":   consts.EmailStatusDel,
			"group_id": 0,
		})
		return err
	}
	// 先删除关联关系
	var ue models.UserEmail
	_, err := session.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", id, ctx.UserID).Delete(&ue)
	if err != nil {
		return err
	}
	// 检查email是否还有人有权限
	var Num num
	_, err = session.Table(&models.UserEmail{}).Select("count(1) as num").Where("email_id=? ", id).Get(&Num)
	if err != nil {
		return err
	}
	if Num.Num == 0 {
		var email models.Email
		_, err = session.Table(&email).Where("id=?", id).Delete(&email)

	}
	return err
}

func DelByUID(ctx *context.Context, ids []int) error {
	session := db.Instance.NewSession()
	defer session.Close()
	for _, id := range ids {
		var ue models.UserEmail
		session.Table("user_email").Where(Eq{"id": id, "user_id": ctx.UserID}).Get(&ue)
		if ue.ID == 0 {
			log.WithContext(ctx).Warn("未找到用户邮件关联记录")
			return nil
		}
		emailId := ue.EmailID

		// 先删除关联关系
		_, err := session.Table(&models.UserEmail{}).Where("id=? and user_id=?", id, ctx.UserID).Delete(&ue)
		if err != nil {
			log.WithContext(ctx).Errorf("删除用户邮件关联失败 ID=%d 错误=%v", id, err)
			session.Rollback()
			return err
		}

		// 检查email是否还有人有权限
		var Num num
		_, err = session.Table(&models.UserEmail{}).Select("count(1) as num").Where("email_id=? ", emailId).Get(&Num)
		if err != nil {
			log.WithContext(ctx).Errorf("查询邮件关联数量失败 邮件ID=%d 错误=%v", emailId, err)
			session.Rollback()
			return err
		}
		if Num.Num == 0 {
			var email models.Email
			_, err = session.Table(&email).Where("id=?", emailId).Delete(&email)
			if err != nil {
				log.WithContext(ctx).Errorf("删除邮件失败 邮件ID=%d 错误=%v", emailId, err)
			}
		}
	}
	session.Commit()
	return nil
}
