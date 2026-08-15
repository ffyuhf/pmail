package list

import (
	"strings"
	"testing"

	"github.com/ffyuhf/pmail/consts"
	"github.com/ffyuhf/pmail/dto"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
)

// TestGenSQL 验证列表查询 SQL 的拼接逻辑（纯函数，不依赖数据库）。
// 移植日期: 20260815 — 配合母项目 genSQL 系统文件夹双条件改进的合并移植新增对照用例，
// 同时锁定现有项目自有修复（GroupId>0 免 status 过滤 / LIKE 通配符转义 / ue_id 查询列）不被回归。
func TestGenSQL(t *testing.T) {
	ctx := &context.Context{UserID: 1}

	t.Run("删除文件夹视图命中status或Deleted系统文件夹", func(t *testing.T) {
		sql, params := genSQL(ctx, false, dto.SearchTag{Type: -1, Status: consts.EmailStatusDel, GroupId: -1}, "", false, 0, 10)
		if !strings.Contains(sql, "(ue.status =? or ue.group_id=?)") {
			t.Fatalf("缺少 status/group_id 双条件: %s", sql)
		}
		hasDeleted := false
		for _, p := range params {
			if p == models.Deleted {
				hasDeleted = true
			}
		}
		if !hasDeleted {
			t.Fatalf("双条件参数应包含 Deleted 系统文件夹值: %v", params)
		}
	})

	t.Run("草稿与垃圾文件夹双条件", func(t *testing.T) {
		for _, status := range []int8{consts.EmailStatusDrafts, consts.EmailStatusJunk} {
			sql, _ := genSQL(ctx, false, dto.SearchTag{Type: -1, Status: status, GroupId: -1}, "", false, 0, 10)
			if !strings.Contains(sql, "(ue.status =? or ue.group_id=?)") {
				t.Fatalf("status=%d 缺少双条件: %s", status, sql)
			}
		}
	})

	t.Run("发件箱视图命中type或Sent系统文件夹", func(t *testing.T) {
		sql, params := genSQL(ctx, false, dto.SearchTag{Type: consts.EmailTypeSend, Status: -1, GroupId: 0}, "", false, 0, 10)
		if !strings.Contains(sql, "(type =? or ue.group_id=?)") {
			t.Fatalf("缺少 type/Sent 双条件: %s", sql)
		}
		if !strings.Contains(sql, "(ue.group_id=? or ue.group_id=0)") {
			t.Fatalf("GroupId=0 且 Status=-1 时应命中 INBOX 或默认文件夹: %s", sql)
		}
		hasSystemFolder := false
		for _, p := range params {
			if p == models.Sent || p == models.INBOX {
				hasSystemFolder = true
			}
		}
		if !hasSystemFolder {
			t.Fatalf("参数应包含 Sent/INBOX 系统文件夹值: %v", params)
		}
	})

	t.Run("自定义分组免status过滤_20260516修复保留", func(t *testing.T) {
		sql, _ := genSQL(ctx, false, dto.SearchTag{Type: -1, Status: -1, GroupId: 5}, "", false, 0, 10)
		if strings.Contains(sql, "ue.status = 0") || strings.Contains(sql, "ue.status != 3") {
			t.Fatalf("自定义分组视图不应限制 status: %s", sql)
		}
		if !strings.Contains(sql, "ue.group_id=?") {
			t.Fatalf("缺少分组过滤条件: %s", sql)
		}
	})

	t.Run("LIKE通配符转义保留_20260610修复保留", func(t *testing.T) {
		sql, params := genSQL(ctx, false, dto.SearchTag{Type: -1, Status: -1, GroupId: -1}, "a%b_c", false, 0, 10)
		if !strings.Contains(sql, "like ?") {
			t.Fatalf("缺少 like 过滤条件: %s", sql)
		}
		escaped := false
		for _, p := range params {
			if s, ok := p.(string); ok && strings.Contains(s, `\%`) && strings.Contains(s, `\_`) {
				escaped = true
			}
		}
		if !escaped {
			t.Fatalf("LIKE 通配符未被转义: %v", params)
		}
	})

	t.Run("ue_id查询列保留", func(t *testing.T) {
		sql, _ := genSQL(ctx, false, dto.SearchTag{Type: -1, Status: -1, GroupId: -1}, "", false, 0, 10)
		if !strings.Contains(sql, "ue.id as ue_id") {
			t.Fatalf("缺少 ue_id 查询列: %s", sql)
		}
	})
}
