package cyno

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var (
	attackFrames          [][]int
	attackHitmarks        = [][]int{{14}, {17}, {13, 22}, {27}}
	attackHitlagHaltFrame = [][]float64{{0.01}, {0.06}, {0, 0.02}, {0.04}}
	attackDefHalt         = [][]bool{{false}, {true}, {false, true}, {true}}
)

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 28) // N1 -> CA
	attackFrames[0][action.ActionAttack] = 15

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 23) // N2 -> CA
	attackFrames[1][action.ActionAttack] = 22

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 27) // N3 -> N4
	attackFrames[2][action.ActionCharge] = 26

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 58) // N4 -> N1
	attackFrames[3][action.ActionCharge] = 500                               // impossible action
}

func (c *char) Attack(p map[string]int) action.ActionInfo {
	if c.StatusIsActive(burstKey) {
		c.NormalHitNum = burstHitNum
		return c.attackB(p) // go to burst mode attacks
	}
	if c.NormalHitNum >= burstHitNum { // this should avoid the panic error
		c.NormalHitNum = normalHitNum
		c.ResetNormalCounter()
	}
	c.NormalHitNum = normalHitNum
	for i, mult := range attack[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Normal %v", c.NormalCounter),
			Mult:               mult[c.TalentLvlAttack()],
			AttackTag:          combat.AttackTagNormal,
			ICDTag:             combat.ICDTagNormalAttack,
			ICDGroup:           combat.ICDGroupDefault,
			Element:            attributes.Physical,
			Durability:         25,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackHitlagHaltFrame[c.NormalCounter][i],
			CanBeDefenseHalted: attackDefHalt[c.NormalCounter][i],
		}

		c.Core.QueueAttack(
			ai,
			c.attackPattern(c.NormalCounter),
			attackHitmarks[c.NormalCounter][i],
			attackHitmarks[c.NormalCounter][i],
		)
	}

	defer c.AdvanceNormalIndex()

	return action.ActionInfo{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[c.NormalCounter][len(attackHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}
}

func (c *char) attackPattern(attackIndex int) combat.AttackPattern {
	switch attackIndex {
	case 0:
		return combat.NewCircleHit(c.Core.Combat.Player(), 1.8, false, combat.TargettableEnemy, combat.TargettableGadget)
	case 1:
		return combat.NewCircleHit(
			c.Core.Combat.Player(),
			1.35,
			false,
			combat.TargettableEnemy,
			combat.TargettableGadget,
		) // supposed to be box x=1.8,z=2.7
	case 2:
		return combat.NewCircleHit(
			c.Core.Combat.Player(),
			1.8,
			false,
			combat.TargettableEnemy,
			combat.TargettableGadget,
		) // both hits supposed to be box x=2.2,z=3.6
	case 3:
		return combat.NewCircleHit(c.Core.Combat.Player(), 2.3, false, combat.TargettableEnemy, combat.TargettableGadget)
	}
	panic("unreachable code")
}

const burstHitNum = 5

var (
	attackBFrames          [][]int
	attackBHitmarks        = [][]int{{12}, {14}, {18}, {5, 14}, {40}}
	attackBHitlagHaltFrame = [][]float64{{0.01}, {0.01}, {0.03}, {0.01, 0.03}, {0.05}}
	attackBDefHalt         = [][]bool{{false}, {false}, {false}, {false, false}, {true}}
)

func init() {
	// NA cancels (burst)
	attackBFrames = make([][]int, burstHitNum)
	attackBFrames[0] = frames.InitNormalCancelSlice(attackBHitmarks[0][0], 28) // N1 -> CA
	attackBFrames[0][action.ActionAttack] = 16

	attackBFrames[1] = frames.InitNormalCancelSlice(attackBHitmarks[1][0], 35) // N2 -> CA
	attackBFrames[1][action.ActionAttack] = 31

	attackBFrames[2] = frames.InitNormalCancelSlice(attackBHitmarks[2][0], 41) // N3 -> N4
	attackBFrames[2][action.ActionCharge] = 39

	attackBFrames[3] = frames.InitNormalCancelSlice(attackBHitmarks[3][0], 36) // N4 -> CA
	attackBFrames[3][action.ActionAttack] = 27

	attackBFrames[4] = frames.InitNormalCancelSlice(attackBHitmarks[4][0], 62) // N5 -> N1
	attackBFrames[4][action.ActionCharge] = 500                                // illegal action
}

func (c *char) attackB(p map[string]int) action.ActionInfo {
	c.tryBurstPPSlide(attackBHitmarks[c.NormalCounter][len(attackBHitmarks[c.NormalCounter])-1])

	for i, mult := range attackB[c.NormalCounter] {
		ai := combat.AttackInfo{
			ActorIndex:         c.Index,
			Abil:               fmt.Sprintf("Pactsworn Pathclearer %v", c.NormalCounter),
			AttackTag:          combat.AttackTagNormal,
			ICDTag:             combat.ICDTagNormalAttack,
			ICDGroup:           combat.ICDGroupDefault,
			Element:            attributes.Electro,
			Durability:         25,
			HitlagFactor:       0.01,
			HitlagHaltFrames:   attackBHitlagHaltFrame[c.NormalCounter][i],
			CanBeDefenseHalted: attackBDefHalt[c.NormalCounter][i],
			Mult:               mult[c.TalentLvlBurst()],
			FlatDmg:            c.Stat(attributes.EM) * 1.5, // this is A4
		}
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, c.attackBPattern(c.NormalCounter), 0, 0)
		}, attackBHitmarks[c.NormalCounter][i])
	}

	defer c.AdvanceNormalIndex()

	return action.ActionInfo{
		Frames:          frames.NewAttackFunc(c.Character, attackBFrames),
		AnimationLength: attackBFrames[c.NormalCounter][action.InvalidAction],
		CanQueueAfter:   attackBHitmarks[c.NormalCounter][len(attackBHitmarks[c.NormalCounter])-1],
		State:           action.NormalAttackState,
	}
}

func (c *char) attackBPattern(attackIndex int) combat.AttackPattern {
	switch attackIndex {
	case 0:
		return combat.NewCircleHit(c.Core.Combat.Player(), 2, false, combat.TargettableEnemy, combat.TargettableGadget)
	case 1:
		return combat.NewCircleHit(c.Core.Combat.Player(), 2, false, combat.TargettableEnemy, combat.TargettableGadget)
	case 2:
		return combat.NewCircleHit(
			c.Core.Combat.Player(),
			3.0,
			false,
			combat.TargettableEnemy,
			combat.TargettableGadget,
		) // supposed to be box x=2.5,z=6.0
	case 3: // both hits are 2.5m radius circles
		return combat.NewCircleHit(
			c.Core.Combat.Player(),
			2.5,
			false,
			combat.TargettableEnemy,
			combat.TargettableGadget,
		)
	case 4:
		return combat.NewCircleHit(c.Core.Combat.Player(), 3.5, false, combat.TargettableEnemy, combat.TargettableGadget)
	}
	panic("unreachable code")
}
