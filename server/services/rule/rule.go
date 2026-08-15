package rule

import (
	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/consts"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto"
	"github.com/ffyuhf/pmail/dto/parsemail"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/services/rule/match"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/send"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"strings"
)

func GetAllRules(ctx *context.Context, userId int) []*dto.Rule {
	var res []*models.Rule
	var err error
	if userId == 0 {
		return nil
	} else {
		err = db.Instance.Where("user_id=?", userId).Decr("sort").Find(&res)
	}

	if err != nil {
		log.WithContext(ctx).Errorf("sqlERror :%v", err)
	}
	var ret []*dto.Rule
	for _, rule := range res {
		ret = append(ret, (&dto.Rule{}).Decode(rule))
	}

	return ret
}

func MatchRule(ctx *context.Context, rule *dto.Rule, email *parsemail.Email) bool {

	for _, r := range rule.Rules {
		var m match.Match

		switch r.Type {
		case match.RuleTypeRegex:
			m = match.NewRegexMatch(r.Field, r.Rule)
		case match.RuleTypeContains:
			m = match.NewContainsMatch(r.Field, r.Rule)
		case match.RuleTypeEq:
			m = match.NewEqualMatch(r.Field, r.Rule)
		}
		if m == nil {
			continue
		}

		if !m.Match(ctx, email) {
			return false
		}
	}

	return true
}

func DoRule(ctx *context.Context, rule *dto.Rule, email *parsemail.Email, user *models.User) {
	log.WithContext(ctx).Debugf("执行规则:%s", rule.Name)

	switch rule.Action {
	case dto.READ:
		if email.MessageId > 0 {
			_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).Cols("is_read").Update(map[string]interface{}{"is_read": 1})
			if err != nil {
				log.WithContext(ctx).Errorf("sqlERror :%v", err)
			}
		}
	case dto.DELETE:
		_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).Cols("status").Update(map[string]interface{}{"status": consts.EmailStatusDel})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}
	case dto.FORWARD:
		if strings.Contains(rule.Params, config.Instance.Domain) {
			log.WithContext(ctx).Errorf("Forward Error! loop forwarding!")
			return
		}
		err := send.Forward(ctx, email, rule.Params, user)
		if err != nil {
			log.WithContext(ctx).Errorf("Forward Error:%v", err)
		}
	case dto.MOVE:
		_, err := db.Instance.Table(&models.UserEmail{}).Where("email_id=? and user_id=?", email.MessageId, rule.UserId).Cols("group_id").Update(map[string]interface{}{"group_id": cast.ToInt(rule.Params)})
		if err != nil {
			log.WithContext(ctx).Errorf("sqlERror :%v", err)
		}
	}

}
