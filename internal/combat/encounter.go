package combat

import (
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
)

const (
	EncounterStart    = "combat-start"
	EncounterEndTurn  = "combat-end-turn"
	EncounterMove     = "combat-move"
	EncounterDash     = "combat-dash"
	EncounterApproach = "combat-approach"
	EncounterRetreat  = "combat-retreat"
)

// EncounterActor is public state for the explicitly known two-actor fixture.
type EncounterActor struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Player        bool      `json:"player"`
	HP            int       `json:"hp,omitempty"`
	MaxHP         int       `json:"max_hp,omitempty"`
	Initiative    int       `json:"initiative"`
	InitiativeDie int       `json:"initiative_die"`
	Speed         int       `json:"speed,omitempty"`
}

// Public returns a detached preview without NPC HP or its unplayed tactic.
// The durable engine state retains these private mechanical inputs.
func (e *Encounter) Public() *Encounter {
	copy := *e
	copy.NPCTactic = ""
	copy.Actors = append([]EncounterActor(nil), e.Actors...)
	for i := range copy.Actors {
		if !copy.Actors[i].Player {
			copy.Actors[i].HP, copy.Actors[i].MaxHP = 0, 0
		}
	}
	return &copy
}

// TurnBudget persists the current actor's remaining turn resources.
type TurnBudget struct {
	SurgeAction    bool `json:"surge_action,omitempty"`
	SurgeUsed      bool `json:"surge_used,omitempty"`
	SpellSlotSpent bool `json:"spell_slot_spent,omitempty"`
	Action         bool `json:"action"`
	BonusAction    bool `json:"bonus_action"`
	Movement       int  `json:"movement"`
}

// Encounter is the supported two-actor fixture, separate from legacy combat.
// Initiative is rolled once. Equal totals put the player first (declared tie choice).
type Encounter struct {
	Bandit    *srd521.BanditProfile `json:"bandit,omitempty"`
	Reaction  *ReactionState        `json:"reaction,omitempty"`
	NPCTactic string                `json:"npc_tactic,omitempty"`
	Space     *Space                `json:"space,omitempty"`
	ID        uuid.UUID             `json:"id"`
	Status    string                `json:"status"`
	Round     int                   `json:"round"`
	Turn      int                   `json:"turn"`
	Actors    []EncounterActor      `json:"actors"`
	Budget    TurnBudget            `json:"budget"`
}

// NewEncounter rolls the ordinary fixture's initiative without model input.
func NewEncounter(player, foe EncounterActor, dexterity int, dice srd521.Dice) (*Encounter, error) {
	p, n := dice.Intn(20)+1, dice.Intn(20)+1
	if p < 1 || p > 20 || n < 1 || n > 20 {
		return nil, fmt.Errorf("invalid initiative dice")
	}
	player.Player, player.Initiative, player.InitiativeDie = true, p+dexterity, p
	foe.Initiative, foe.InitiativeDie = n+foe.Initiative, n
	actors := []EncounterActor{player, foe}
	if foe.Initiative > player.Initiative {
		actors[0], actors[1] = foe, player
	}
	return &Encounter{Reaction: &ReactionState{Available: true}, ID: uuid.New(), Status: "active", Round: 1, Actors: actors, Budget: TurnBudget{Action: true, BonusAction: true, Movement: actorSpeed(actors[0])}}, nil
}

// Active returns the actor whose turn owns the budget.
func (e *Encounter) Active() *EncounterActor { return &e.Actors[e.Turn] }

// Spend validates and consumes one supported resource. A move spends five feet;
// scene/range adjudication remains a separate boundary.
func (e *Encounter) Spend(actor uuid.UUID, kind string) error {
	if e.Status != "active" || e.Active().ID != actor || e.Active().HP <= 0 {
		return fmt.Errorf("not this actor's turn")
	}
	switch kind {
	case srd521.ActionSurge:
		if e.Budget.SurgeUsed {
			return fmt.Errorf("Action Surge already used this turn")
		}
		e.Budget.SurgeUsed, e.Budget.SurgeAction = true, true
	case srd521.SpellCast, srd521.MartialAttack, EncounterDash:
		if e.Budget.Action {
			e.Budget.Action = false
		} else if kind != srd521.SpellCast && e.Budget.SurgeAction {
			e.Budget.SurgeAction = false
		} else {
			return fmt.Errorf("no eligible action remains")
		}
		if kind == EncounterDash {
			e.Budget.Movement += actorSpeed(*e.Active())
		}
	case srd521.ConsumableUse:
		if !e.Budget.BonusAction {
			return fmt.Errorf("bonus action already spent")
		}
		e.Budget.BonusAction = false
	case EncounterMove, EncounterApproach, EncounterRetreat:
		if e.Budget.Movement < 5 {
			return fmt.Errorf("insufficient movement")
		}
		if kind != EncounterMove {
			if e.Space == nil {
				return fmt.Errorf("Clarify spatial facts before moving.")
			}
			space := *e.Space
			if err := space.Move(kind); err != nil {
				return err
			}
			e.Space = &space
		}
		e.Budget.Movement -= 5
	case EncounterEndTurn:
	default:
		return fmt.Errorf("unsupported encounter action")
	}
	return nil
}

// Advance ends a turn, resets its budget, and increments the round on wrap.
func (e *Encounter) Advance() {
	if e.Status != "active" {
		return
	}
	e.Turn++
	if e.Turn == len(e.Actors) {
		e.Turn = 0
		e.Round++
	}
	e.Budget = TurnBudget{Action: true, BonusAction: true, Movement: actorSpeed(*e.Active())}
	if e.Active().Player && e.Reaction != nil {
		e.Reaction.Available = true
	}
}

// SetHP synchronizes a committed actor change and finishes a defeated fixture.
func (e *Encounter) SetHP(id uuid.UUID, hp int) {
	for i := range e.Actors {
		if e.Actors[i].ID == id {
			e.Actors[i].HP = hp
		}
	}
	for _, a := range e.Actors {
		if a.HP <= 0 {
			e.Status = "completed"
		}
	}
}

// SpendSpell enforces the SRD one-spell-slot-per-turn limit independently of
// action versus Bonus Action costs. It resets with the actor's turn budget.
func (e *Encounter) SpendSpell(actor uuid.UUID, bonus bool) error {
	if e.Budget.SpellSlotSpent {
		return fmt.Errorf("one spell slot can be spent per turn")
	}
	kind := srd521.SpellCast
	if bonus {
		kind = srd521.ConsumableUse
	}
	if err := e.Spend(actor, kind); err != nil {
		return err
	}
	e.Budget.SpellSlotSpent = true
	return nil
}

// actorSpeed keeps historical fixtures at 30 feet while honoring pinned actors.
func actorSpeed(actor EncounterActor) int {
	if actor.Speed > 0 {
		return actor.Speed
	}
	return 30
}
