package imap_server

import (
	"github.com/ffyuhf/pmail/services/list"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// Search 实现了 IMAP SEARCH 命令，支持完整的搜索条件
// 支持：UID、序列号、日期过滤、头部搜索、正文/全文搜索、
// 标志过滤、大小过滤以及逻辑组合（NOT、OR）
func (s *serverSession) Search(kind imapserver.NumKind, criteria *imap.SearchCriteria, options *imap.SearchOptions) (*imap.SearchData, error) {
	log.WithContext(s.ctx).Debugf("IMAP SEARCH: mailbox=%s, kind=%v, criteria=%+v", s.currentMailbox, kind, criteria)

	// 使用新的综合搜索函数
	retList, err := list.SearchEmails(s.ctx, s.currentMailbox, criteria)
	if err != nil {
		log.WithContext(s.ctx).Errorf("IMAP SEARCH error: %v", err)
		return nil, err
	}

	ret := &imap.SearchData{}

	if kind == imapserver.NumKindSeq {
		// 返回序列号
		idList := imap.SeqSet{}
		for _, data := range retList {
			idList = append(idList, imap.SeqRange{
				Start: cast.ToUint32(data.SerialNumber),
				Stop:  cast.ToUint32(data.SerialNumber),
			})
		}
		ret.All = idList
		ret.Count = uint32(len(retList))

		// 处理 ESEARCH 选项
		if options != nil {
			if options.ReturnMin && len(retList) > 0 {
				ret.Min = cast.ToUint32(retList[0].SerialNumber)
				for _, data := range retList {
					if cast.ToUint32(data.SerialNumber) < ret.Min {
						ret.Min = cast.ToUint32(data.SerialNumber)
					}
				}
			}
			if options.ReturnMax && len(retList) > 0 {
				ret.Max = cast.ToUint32(retList[0].SerialNumber)
				for _, data := range retList {
					if cast.ToUint32(data.SerialNumber) > ret.Max {
						ret.Max = cast.ToUint32(data.SerialNumber)
					}
				}
			}
		}
	} else {
		// 返回 UID
		idList := imap.UIDSet{}
		for _, data := range retList {
			idList = append(idList, imap.UIDRange{
				Start: imap.UID(data.ID),
				Stop:  imap.UID(data.ID),
			})
		}
		ret.UID = true
		ret.All = idList
		ret.Count = uint32(len(retList))

		// 处理 ESEARCH 选项
		if options != nil {
			if options.ReturnMin && len(retList) > 0 {
				ret.Min = cast.ToUint32(retList[0].ID)
				for _, data := range retList {
					if cast.ToUint32(data.ID) < ret.Min {
						ret.Min = cast.ToUint32(data.ID)
					}
				}
			}
			if options.ReturnMax && len(retList) > 0 {
				ret.Max = cast.ToUint32(retList[0].ID)
				for _, data := range retList {
					if cast.ToUint32(data.ID) > ret.Max {
						ret.Max = cast.ToUint32(data.ID)
					}
				}
			}
		}
	}

	log.WithContext(s.ctx).Debugf("IMAP SEARCH result: count=%d", ret.Count)
	return ret, nil
}
