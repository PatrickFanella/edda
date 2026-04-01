package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type stubAddItemStore struct {
	lastPlayerID    uuid.UUID
	lastName        string
	lastDescription string
	lastType        string
	lastRarity      string
	lastQuantity    int

	itemID uuid.UUID
	err    error
}

func (s *stubAddItemStore) CreatePlayerItem(_ context.Context, playerCharacterID uuid.UUID, name, description, itemType, rarity string, quantity int) (uuid.UUID, error) {
	if s.err != nil {
		return uuid.Nil, s.err
	}
	s.lastPlayerID = playerCharacterID
	s.lastName = name
	s.lastDescription = description
	s.lastType = itemType
	s.lastRarity = rarity
	s.lastQuantity = quantity
	if s.itemID == uuid.Nil {
		s.itemID = uuid.New()
	}
	return s.itemID, nil
}

type stubRemoveItemStore struct {
	items map[uuid.UUID]*PlayerItem

	lastUpdatedID       uuid.UUID
	lastUpdatedQuantity int
	lastDeletedID       uuid.UUID

	getErr    error
	updateErr error
	deleteErr error
}

func (s *stubRemoveItemStore) GetPlayerItemByID(_ context.Context, itemID uuid.UUID) (*PlayerItem, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	item := s.items[itemID]
	if item == nil {
		return nil, nil
	}
	copied := *item
	return &copied, nil
}

func (s *stubRemoveItemStore) UpdateItemQuantity(_ context.Context, itemID uuid.UUID, quantity int) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.lastUpdatedID = itemID
	s.lastUpdatedQuantity = quantity
	return nil
}

type stubCreateItemStore struct {
	lastPlayerID    uuid.UUID
	lastName        string
	lastDescription string
	lastType        string
	lastRarity      string
	lastProperties  map[string]any

	itemID uuid.UUID
	err    error
}

func (s *stubCreateItemStore) CreateGeneratedItem(_ context.Context, playerCharacterID uuid.UUID, name, description, itemType, rarity string, properties map[string]any) (uuid.UUID, error) {
	if s.err != nil {
		return uuid.Nil, s.err
	}
	s.lastPlayerID = playerCharacterID
	s.lastName = name
	s.lastDescription = description
	s.lastType = itemType
	s.lastRarity = rarity
	s.lastProperties = properties
	if s.itemID == uuid.Nil {
		s.itemID = uuid.New()
	}
	return s.itemID, nil
}

type stubModifyItemStore struct {
	items map[uuid.UUID]*PlayerItem

	lastUpdatedID         uuid.UUID
	lastUpdatedProperties map[string]any

	getErr    error
	updateErr error
}

func (s *stubModifyItemStore) GetPlayerItemByID(_ context.Context, itemID uuid.UUID) (*PlayerItem, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	item := s.items[itemID]
	if item == nil {
		return nil, nil
	}
	copied := *item
	return &copied, nil
}

func (s *stubModifyItemStore) UpdatePlayerItemProperties(_ context.Context, itemID uuid.UUID, properties map[string]any) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.lastUpdatedID = itemID
	s.lastUpdatedProperties = properties
	return nil
}

func (s *stubRemoveItemStore) DeleteItem(_ context.Context, itemID uuid.UUID) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	s.lastDeletedID = itemID
	return nil
}

func TestRegisterAddItem(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterAddItem(reg, &stubAddItemStore{}); err != nil {
		t.Fatalf("register add_item: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != addItemToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, addItemToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 3 {
		t.Fatalf("required schema length = %d, want 3", len(required))
	}
}

func TestRegisterRemoveItem(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterRemoveItem(reg, &stubRemoveItemStore{}); err != nil {
		t.Fatalf("register remove_item: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != removeItemToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, removeItemToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 1 || required[0] != "item_id" {
		t.Fatalf("required schema = %#v, want [item_id]", required)
	}
}

func TestRegisterCreateItem(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterCreateItem(reg, &stubCreateItemStore{}); err != nil {
		t.Fatalf("register create_item: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != createItemToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, createItemToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 5 {
		t.Fatalf("required schema length = %d, want 5", len(required))
	}
}

func TestRegisterModifyItem(t *testing.T) {
	reg := NewRegistry()
	if err := RegisterModifyItem(reg, &stubModifyItemStore{}); err != nil {
		t.Fatalf("register modify_item: %v", err)
	}
	registered := reg.List()
	if len(registered) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(registered))
	}
	if registered[0].Name != modifyItemToolName {
		t.Fatalf("tool name = %q, want %q", registered[0].Name, modifyItemToolName)
	}
	required, ok := registered[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", registered[0].Parameters["required"])
	}
	if len(required) != 2 {
		t.Fatalf("required schema length = %d, want 2", len(required))
	}
}

func TestAddItemHandleDefaultQuantity(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddItemStore{itemID: uuid.New()}
	h := NewAddItemHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"name":        "Potion",
		"description": "Restores health",
		"item_type":   "consumable",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if store.lastPlayerID != playerID {
		t.Fatalf("stored player id = %s, want %s", store.lastPlayerID, playerID)
	}
	if store.lastQuantity != 1 {
		t.Fatalf("stored quantity = %d, want 1", store.lastQuantity)
	}
	if got.Data["quantity"] != 1 {
		t.Fatalf("result quantity = %v, want 1", got.Data["quantity"])
	}
}

func TestAddItemHandleInvalidInputs(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("missing context player id", func(t *testing.T) {
		h := NewAddItemHandler(&stubAddItemStore{})
		_, err := h.Handle(context.Background(), map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "consumable",
		})
		if err == nil || !strings.Contains(err.Error(), "requires current player character id in context") {
			t.Fatalf("error = %v, want context player id error", err)
		}
	})

	t.Run("invalid item type", func(t *testing.T) {
		h := NewAddItemHandler(&stubAddItemStore{})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "invalid",
		})
		if err == nil || !strings.Contains(err.Error(), "item_type must be one of") {
			t.Fatalf("error = %v, want item_type validation error", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		h := NewAddItemHandler(&stubAddItemStore{})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "consumable",
			"quantity":    0,
		})
		if err == nil || !strings.Contains(err.Error(), "quantity must be greater than 0") {
			t.Fatalf("error = %v, want quantity validation error", err)
		}
	})

	t.Run("store error wrapped", func(t *testing.T) {
		h := NewAddItemHandler(&stubAddItemStore{err: errors.New("db down")})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "consumable",
		})
		if err == nil || !strings.Contains(err.Error(), "create item: db down") {
			t.Fatalf("error = %v, want wrapped store error", err)
		}
	})
}

func TestRemoveItemHandleDecrementAndDelete(t *testing.T) {
	playerID := uuid.New()
	itemID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("decrement quantity", func(t *testing.T) {
		store := &stubRemoveItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Arrow", Quantity: 5},
			},
		}
		h := NewRemoveItemHandler(store)
		got, err := h.Handle(ctx, map[string]any{
			"item_id":  itemID.String(),
			"quantity": 2,
		})
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if store.lastUpdatedID != itemID || store.lastUpdatedQuantity != 3 {
			t.Fatalf("updated id/qty = %s/%d, want %s/3", store.lastUpdatedID, store.lastUpdatedQuantity, itemID)
		}
		if store.lastDeletedID != uuid.Nil {
			t.Fatalf("unexpected delete id = %s", store.lastDeletedID)
		}
		if got.Data["remaining_quantity"] != 3 {
			t.Fatalf("remaining_quantity = %v, want 3", got.Data["remaining_quantity"])
		}
	})

	t.Run("default quantity removes all and deletes", func(t *testing.T) {
		store := &stubRemoveItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Arrow", Quantity: 5},
			},
		}
		h := NewRemoveItemHandler(store)
		got, err := h.Handle(ctx, map[string]any{
			"item_id": itemID.String(),
		})
		if err != nil {
			t.Fatalf("Handle: %v", err)
		}
		if store.lastDeletedID != itemID {
			t.Fatalf("deleted id = %s, want %s", store.lastDeletedID, itemID)
		}
		if store.lastUpdatedID != uuid.Nil {
			t.Fatalf("unexpected update id = %s", store.lastUpdatedID)
		}
		if got.Data["deleted"] != true {
			t.Fatalf("deleted flag = %v, want true", got.Data["deleted"])
		}
	})
}

func TestRemoveItemHandleInvalidInputs(t *testing.T) {
	playerID := uuid.New()
	otherPlayerID := uuid.New()
	itemID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("missing context player id", func(t *testing.T) {
		h := NewRemoveItemHandler(&stubRemoveItemStore{})
		_, err := h.Handle(context.Background(), map[string]any{
			"item_id": itemID.String(),
		})
		if err == nil || !strings.Contains(err.Error(), "requires current player character id in context") {
			t.Fatalf("error = %v, want context player id error", err)
		}
	})

	t.Run("item not found", func(t *testing.T) {
		h := NewRemoveItemHandler(&stubRemoveItemStore{items: map[uuid.UUID]*PlayerItem{}})
		_, err := h.Handle(ctx, map[string]any{"item_id": itemID.String()})
		if err == nil || !strings.Contains(err.Error(), "does not reference an existing item") {
			t.Fatalf("error = %v, want missing item error", err)
		}
	})

	t.Run("item belongs to other player", func(t *testing.T) {
		h := NewRemoveItemHandler(&stubRemoveItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: otherPlayerID, Name: "Ring", Quantity: 1},
			},
		})
		_, err := h.Handle(ctx, map[string]any{"item_id": itemID.String()})
		if err == nil || !strings.Contains(err.Error(), "does not belong to current player") {
			t.Fatalf("error = %v, want ownership error", err)
		}
	})

	t.Run("quantity exceeds", func(t *testing.T) {
		h := NewRemoveItemHandler(&stubRemoveItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Ring", Quantity: 1},
			},
		})
		_, err := h.Handle(ctx, map[string]any{
			"item_id":  itemID.String(),
			"quantity": 2,
		})
		if err == nil || !strings.Contains(err.Error(), "quantity exceeds item quantity") {
			t.Fatalf("error = %v, want quantity exceeds error", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		h := NewRemoveItemHandler(&stubRemoveItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Ring", Quantity: 1},
			},
		})
		_, err := h.Handle(ctx, map[string]any{
			"item_id":  itemID.String(),
			"quantity": 0,
		})
		if err == nil || !strings.Contains(err.Error(), "quantity must be greater than 0") {
			t.Fatalf("error = %v, want invalid quantity error", err)
		}
	})

	t.Run("wrapped store errors", func(t *testing.T) {
		hGet := NewRemoveItemHandler(&stubRemoveItemStore{getErr: errors.New("read fail")})
		_, err := hGet.Handle(ctx, map[string]any{"item_id": itemID.String()})
		if err == nil || !strings.Contains(err.Error(), "get item: read fail") {
			t.Fatalf("error = %v, want wrapped get error", err)
		}

		hUpd := NewRemoveItemHandler(&stubRemoveItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Ring", Quantity: 2},
			},
			updateErr: errors.New("write fail"),
		})
		_, err = hUpd.Handle(ctx, map[string]any{"item_id": itemID.String(), "quantity": 1})
		if err == nil || !strings.Contains(err.Error(), "update item quantity: write fail") {
			t.Fatalf("error = %v, want wrapped update error", err)
		}

		hDel := NewRemoveItemHandler(&stubRemoveItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Ring", Quantity: 1},
			},
			deleteErr: errors.New("delete fail"),
		})
		_, err = hDel.Handle(ctx, map[string]any{"item_id": itemID.String()})
		if err == nil || !strings.Contains(err.Error(), "delete item: delete fail") {
			t.Fatalf("error = %v, want wrapped delete error", err)
		}
	})
}

func TestCreateItemHandle(t *testing.T) {
	playerID := uuid.New()
	store := &stubCreateItemStore{itemID: uuid.New()}
	h := NewCreateItemHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"name":        "Stormbreaker",
		"description": "A runed war axe humming with thunder.",
		"item_type":   "weapon",
		"rarity":      "epic",
		"properties": map[string]any{
			"damage":  "2d8+3",
			"effects": []any{"lightning", "stun"},
			"weight":  6.5,
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if store.lastPlayerID != playerID {
		t.Fatalf("stored player id = %s, want %s", store.lastPlayerID, playerID)
	}
	if store.lastRarity != "epic" {
		t.Fatalf("stored rarity = %q, want epic", store.lastRarity)
	}
	if got.Data["formatted_description"] == "" {
		t.Fatal("formatted_description should not be empty")
	}
}

func TestCreateItemHandleInvalidInputs(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("missing context player id", func(t *testing.T) {
		h := NewCreateItemHandler(&stubCreateItemStore{})
		_, err := h.Handle(context.Background(), map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "consumable",
			"rarity":      "common",
			"properties":  map[string]any{"effects": "heal"},
		})
		if err == nil || !strings.Contains(err.Error(), "requires current player character id in context") {
			t.Fatalf("error = %v, want context player id error", err)
		}
	})

	t.Run("invalid rarity", func(t *testing.T) {
		h := NewCreateItemHandler(&stubCreateItemStore{})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "consumable",
			"rarity":      "mythic",
			"properties":  map[string]any{"effects": "heal"},
		})
		if err == nil || !strings.Contains(err.Error(), "rarity must be one of") {
			t.Fatalf("error = %v, want rarity validation error", err)
		}
	})

	t.Run("unsupported properties", func(t *testing.T) {
		h := NewCreateItemHandler(&stubCreateItemStore{})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "consumable",
			"rarity":      "common",
			"properties":  map[string]any{"speed": 10},
		})
		if err == nil || !strings.Contains(err.Error(), "supports only") {
			t.Fatalf("error = %v, want property validation error", err)
		}
	})

	t.Run("store error wrapped", func(t *testing.T) {
		h := NewCreateItemHandler(&stubCreateItemStore{err: errors.New("db down")})
		_, err := h.Handle(ctx, map[string]any{
			"name":        "Potion",
			"description": "Restores health",
			"item_type":   "consumable",
			"rarity":      "common",
			"properties":  map[string]any{"effects": "heal"},
		})
		if err == nil || !strings.Contains(err.Error(), "create item: db down") {
			t.Fatalf("error = %v, want wrapped store error", err)
		}
	})
}

func TestModifyItemHandle(t *testing.T) {
	playerID := uuid.New()
	itemID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	store := &stubModifyItemStore{
		items: map[uuid.UUID]*PlayerItem{
			itemID: {
				ID:                itemID,
				PlayerCharacterID: playerID,
				Name:              "Wand of Sparks",
				Description:       "Crackles with energy.",
				ItemType:          "weapon",
				Rarity:            "rare",
				Properties: map[string]any{
					"charges": 3,
					"effects": "spark",
				},
			},
		},
	}
	h := NewModifyItemHandler(store)

	got, err := h.Handle(ctx, map[string]any{
		"item_id": itemID.String(),
		"properties": map[string]any{
			"charges": 2,
			"damage":  "1d6",
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if store.lastUpdatedID != itemID {
		t.Fatalf("updated id = %s, want %s", store.lastUpdatedID, itemID)
	}
	if store.lastUpdatedProperties["charges"] != 2 {
		t.Fatalf("updated charges = %v, want 2", store.lastUpdatedProperties["charges"])
	}
	if store.lastUpdatedProperties["effects"] != "spark" {
		t.Fatalf("existing effect should remain merged, got %v", store.lastUpdatedProperties["effects"])
	}
	if got.Data["formatted_description"] == "" {
		t.Fatal("formatted_description should not be empty")
	}
}

func TestModifyItemHandleInvalidInputs(t *testing.T) {
	playerID := uuid.New()
	otherPlayerID := uuid.New()
	itemID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	t.Run("missing context player id", func(t *testing.T) {
		h := NewModifyItemHandler(&stubModifyItemStore{})
		_, err := h.Handle(context.Background(), map[string]any{
			"item_id":    itemID.String(),
			"properties": map[string]any{"charges": 2},
		})
		if err == nil || !strings.Contains(err.Error(), "requires current player character id in context") {
			t.Fatalf("error = %v, want context player id error", err)
		}
	})

	t.Run("item not found", func(t *testing.T) {
		h := NewModifyItemHandler(&stubModifyItemStore{items: map[uuid.UUID]*PlayerItem{}})
		_, err := h.Handle(ctx, map[string]any{"item_id": itemID.String(), "properties": map[string]any{"charges": 2}})
		if err == nil || !strings.Contains(err.Error(), "does not reference an existing item") {
			t.Fatalf("error = %v, want missing item error", err)
		}
	})

	t.Run("item belongs to other player", func(t *testing.T) {
		h := NewModifyItemHandler(&stubModifyItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: otherPlayerID, Name: "Ring", ItemType: "armor", Rarity: "common", Quantity: 1},
			},
		})
		_, err := h.Handle(ctx, map[string]any{"item_id": itemID.String(), "properties": map[string]any{"charges": 2}})
		if err == nil || !strings.Contains(err.Error(), "does not belong to current player") {
			t.Fatalf("error = %v, want ownership error", err)
		}
	})

	t.Run("unsupported properties", func(t *testing.T) {
		h := NewModifyItemHandler(&stubModifyItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Ring", ItemType: "armor", Rarity: "common"},
			},
		})
		_, err := h.Handle(ctx, map[string]any{"item_id": itemID.String(), "properties": map[string]any{"speed": 3}})
		if err == nil || !strings.Contains(err.Error(), "supports only") {
			t.Fatalf("error = %v, want property validation error", err)
		}
	})

	t.Run("wrapped store errors", func(t *testing.T) {
		hGet := NewModifyItemHandler(&stubModifyItemStore{getErr: errors.New("read fail")})
		_, err := hGet.Handle(ctx, map[string]any{"item_id": itemID.String(), "properties": map[string]any{"charges": 2}})
		if err == nil || !strings.Contains(err.Error(), "get item: read fail") {
			t.Fatalf("error = %v, want wrapped get error", err)
		}

		hUpd := NewModifyItemHandler(&stubModifyItemStore{
			items: map[uuid.UUID]*PlayerItem{
				itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Ring", ItemType: "armor", Rarity: "common"},
			},
			updateErr: errors.New("write fail"),
		})
		_, err = hUpd.Handle(ctx, map[string]any{"item_id": itemID.String(), "properties": map[string]any{"charges": 2}})
		if err == nil || !strings.Contains(err.Error(), "update item properties: write fail") {
			t.Fatalf("error = %v, want wrapped update error", err)
		}
	})
}

func TestAddItemHandleExplicitQuantity(t *testing.T) {
	playerID := uuid.New()
	store := &stubAddItemStore{itemID: uuid.New()}
	h := NewAddItemHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	got, err := h.Handle(ctx, map[string]any{
		"name":        "Arrow",
		"description": "A sharp arrow.",
		"item_type":   "misc",
		"quantity":    5,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if store.lastQuantity != 5 {
		t.Fatalf("stored quantity = %d, want 5", store.lastQuantity)
	}
	if got.Data["quantity"] != 5 {
		t.Fatalf("result quantity = %v, want 5", got.Data["quantity"])
	}
}

func TestAddItemHandleAllItemTypes(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	for _, itemType := range []string{"weapon", "armor", "consumable", "quest", "misc"} {
		itemType := itemType
		t.Run(itemType, func(t *testing.T) {
			store := &stubAddItemStore{itemID: uuid.New()}
			h := NewAddItemHandler(store)
			_, err := h.Handle(ctx, map[string]any{
				"name":        "Test Item",
				"description": "A test item.",
				"item_type":   itemType,
			})
			if err != nil {
				t.Fatalf("Handle(%s): %v", itemType, err)
			}
			if store.lastType != itemType {
				t.Fatalf("stored type = %q, want %q", store.lastType, itemType)
			}
		})
	}
}

func TestAddItemHandleMissingRequiredFields(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	cases := []struct {
		name    string
		args    map[string]any
		wantErr string
	}{
		{
			name:    "missing name",
			args:    map[string]any{"description": "d", "item_type": "misc"},
			wantErr: "name is required",
		},
		{
			name:    "missing description",
			args:    map[string]any{"name": "n", "item_type": "misc"},
			wantErr: "description is required",
		},
		{
			name:    "missing item_type",
			args:    map[string]any{"name": "n", "description": "d"},
			wantErr: "item_type is required",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := NewAddItemHandler(&stubAddItemStore{})
			_, err := h.Handle(ctx, tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestCreateItemHandleAllRarities(t *testing.T) {
	playerID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	for _, rarity := range []string{"common", "uncommon", "rare", "epic", "legendary"} {
		rarity := rarity
		t.Run(rarity, func(t *testing.T) {
			store := &stubCreateItemStore{itemID: uuid.New()}
			h := NewCreateItemHandler(store)
			_, err := h.Handle(ctx, map[string]any{
				"name":        "Test Item",
				"description": "A test item.",
				"item_type":   "misc",
				"rarity":      rarity,
				"properties":  map[string]any{},
			})
			if err != nil {
				t.Fatalf("Handle(%s): %v", rarity, err)
			}
			if store.lastRarity != rarity {
				t.Fatalf("stored rarity = %q, want %q", store.lastRarity, rarity)
			}
		})
	}
}

func TestCreateItemHandleEmptyProperties(t *testing.T) {
	playerID := uuid.New()
	store := &stubCreateItemStore{itemID: uuid.New()}
	h := NewCreateItemHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"name":        "Empty Props Item",
		"description": "An item with no properties.",
		"item_type":   "misc",
		"rarity":      "common",
		"properties":  map[string]any{},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
}

func TestCreateItemHandleAllPropertyKeys(t *testing.T) {
	playerID := uuid.New()
	store := &stubCreateItemStore{itemID: uuid.New()}
	h := NewCreateItemHandler(store)
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)

	_, err := h.Handle(ctx, map[string]any{
		"name":        "Full Props Item",
		"description": "An item with all supported properties.",
		"item_type":   "weapon",
		"rarity":      "rare",
		"properties": map[string]any{
			"effects": []any{"fire"},
			"damage":  "1d8",
			"armor":   5,
			"charges": 3,
			"weight":  2.5,
		},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
}

func TestModifyItemHandleEmptyProperties(t *testing.T) {
	playerID := uuid.New()
	itemID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	store := &stubModifyItemStore{
		items: map[uuid.UUID]*PlayerItem{
			itemID: {
				ID:                itemID,
				PlayerCharacterID: playerID,
				Name:              "Sword",
				ItemType:          "weapon",
				Rarity:            "common",
				Properties:        map[string]any{"charges": 3},
			},
		},
	}
	h := NewModifyItemHandler(store)

	_, err := h.Handle(ctx, map[string]any{
		"item_id":    itemID.String(),
		"properties": map[string]any{},
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	// Existing properties should be preserved when empty update is applied.
	if store.lastUpdatedProperties["charges"] != 3 {
		t.Fatalf("charges should be preserved, got %v", store.lastUpdatedProperties["charges"])
	}
}

func TestRemoveItemHandleExactQuantityDeletes(t *testing.T) {
	playerID := uuid.New()
	itemID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	store := &stubRemoveItemStore{
		items: map[uuid.UUID]*PlayerItem{
			itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Bomb", Quantity: 5},
		},
	}
	h := NewRemoveItemHandler(store)

	got, err := h.Handle(ctx, map[string]any{
		"item_id":  itemID.String(),
		"quantity": 5,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if store.lastDeletedID != itemID {
		t.Fatalf("deleted id = %s, want %s", store.lastDeletedID, itemID)
	}
	if store.lastUpdatedID != uuid.Nil {
		t.Fatalf("unexpected update id = %s", store.lastUpdatedID)
	}
	if got.Data["deleted"] != true {
		t.Fatalf("deleted flag = %v, want true", got.Data["deleted"])
	}
}

func TestRemoveItemHandleNegativeQuantity(t *testing.T) {
	playerID := uuid.New()
	itemID := uuid.New()
	ctx := WithCurrentPlayerCharacterID(context.Background(), playerID)
	store := &stubRemoveItemStore{
		items: map[uuid.UUID]*PlayerItem{
			itemID: {ID: itemID, PlayerCharacterID: playerID, Name: "Bomb", Quantity: 5},
		},
	}
	h := NewRemoveItemHandler(store)

	_, err := h.Handle(ctx, map[string]any{
		"item_id":  itemID.String(),
		"quantity": -1,
	})
	if err == nil || !strings.Contains(err.Error(), "quantity must be greater than 0") {
		t.Fatalf("error = %v, want quantity must be greater than 0", err)
	}
}

var _ AddItemStore = (*stubAddItemStore)(nil)
var _ RemoveItemStore = (*stubRemoveItemStore)(nil)
var _ CreateItemStore = (*stubCreateItemStore)(nil)
var _ ModifyItemStore = (*stubModifyItemStore)(nil)
