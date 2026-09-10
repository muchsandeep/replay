package replay

import (
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

type RecordPlayerHandler struct {
	player.NopHandler

	r *Recorder
}

func NewRecordPlayerHandler(r *Recorder) *RecordPlayerHandler {
	return &RecordPlayerHandler{r: r}
}

func (h *RecordPlayerHandler) HandleMove(ctx *player.Context, pos mgl64.Vec3, rot cube.Rotation) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerMovement(ctx.Player(), pos, rot)
}

func (h *RecordPlayerHandler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	if ctx.Cancelled() {
		return
	}
	// TODO: don't use movement
	h.r.PushPlayerMovement(ctx.Player(), pos, ctx.Player().Rotation())
}

func (h *RecordPlayerHandler) HandleToggleSneak(ctx *player.Context, sneaking bool) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerSneaking(ctx.Player(), sneaking)
}

func (h *RecordPlayerHandler) HandleHeldSlotChange(ctx *player.Context, _, to int) {
	if ctx.Cancelled() {
		return
	}
	_, offHand := ctx.Player().HeldItems()
	mainHand, _ := ctx.Player().Inventory().Item(to)
	h.r.PushPlayerHandChange(ctx.Player(), mainHand, offHand)
	h.r.PushPlayerUsingItem(ctx.Player(), false)
}

func (h *RecordPlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlaceBlock(pos, b)
	if !hasSwingArmHandler {
		h.r.PushPlayerSwingArm(ctx.Player())
	}
}

func (h *RecordPlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, _ *[]item.Stack, _ *int) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushBreakBlock(pos)
	if !hasSwingArmHandler {
		h.r.PushPlayerSwingArm(ctx.Player())
	}
}

func (h *RecordPlayerHandler) HandleItemConsume(ctx *player.Context, _ item.Stack) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerUsingItem(ctx.Player(), false)
}

func (h *RecordPlayerHandler) HandleItemRelease(ctx *player.Context, _ item.Stack, _ time.Duration) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerUsingItem(ctx.Player(), false)
}

func (h *RecordPlayerHandler) HandleItemUse(ctx *player.Context) {
	if ctx.Cancelled() {
		return
	}
	mainHand, _ := ctx.Player().HeldItems()
	switch mainHand.Item().(type) {
	case item.Releasable:
		if !player_canRelease(ctx.Player()) {
			return
		}
		h.r.PushPlayerUsingItem(ctx.Player(), true)
	case item.Consumable:
		h.r.PushPlayerEating(ctx.Player())
		h.r.PushPlayerUsingItem(ctx.Player(), true)
	default:
		// handle other usable item, like crossbow
	}
}

func (h *RecordPlayerHandler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushSkinChange(ctx.Player(), *skin)
}

func (h *RecordPlayerHandler) HandleHurt(ctx *player.Context, _ *float64, _ bool, _ *time.Duration, _ world.DamageSource) {
	if ctx.Cancelled() {
		return
	}
	h.r.PushPlayerHurt(ctx.Player())
}

func (h *RecordPlayerHandler) HandlePunchAir(ctx *player.Context) {
	if ctx.Cancelled() || hasSwingArmHandler {
		return
	}
	h.r.PushPlayerSwingArm(ctx.Player())
}

func (h *RecordPlayerHandler) HandleItemUseOnEntity(ctx *player.Context, _ world.Entity) {}

func (h *RecordPlayerHandler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, _ cube.Face, _ mgl64.Vec3) {
	if ctx.Cancelled() || hasSwingArmHandler {
		return
	}
	b := ctx.Player().Tx().Block(pos)
	if _, ok := b.(block.Activatable); ok {
		h.r.PushPlayerSwingArm(ctx.Player())
		return
	}

	mainHand, _ := ctx.Player().HeldItems()
	if _, ok := mainHand.Item().(item.UsableOnBlock); ok {
		h.r.PushPlayerSwingArm(ctx.Player())
		return
	}
}

func (h *RecordPlayerHandler) HandleAttackEntity(ctx *player.Context, _ world.Entity, _, _ *float64, _ *bool) {
	if ctx.Cancelled() || hasSwingArmHandler {
		return
	}
	h.r.PushPlayerSwingArm(ctx.Player())
}

func (h *RecordPlayerHandler) HandleSwingArm(p *player.Player) {
	h.r.PushPlayerSwingArm(p)
}

var hasSwingArmHandler = false

func init() {
	if _, ok := toAny(player.NopHandler{}).(interface{ HandleSwingArm(p *player.Player) }); ok {
		hasSwingArmHandler = true
	}
}
