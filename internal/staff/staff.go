package staff

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Staff はアクティブレコードパターンで実装されたスタッフ管理。
//
// ドメインモデルパターン（domain/）との違い:
//   - DB 操作とバリデーションが同じパッケージに同居
//   - Repository IF を domain 層に分離しない
//   - usecase 層がない（handler → Store 直結）
//   - 業務ルールが少ない（バリデーション程度）
//
// 第10章: 補完領域 → アクティブレコード → レイヤードアーキテクチャ
type Staff struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	ShiftType string `json:"shift_type"`
}

// Validate はバリデーション。業務ルールはこれだけ。
func (s *Staff) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("名前は必須です")
	}
	if !strings.Contains(s.Phone, "-") || len(s.Phone) < 10 {
		return errors.New("電話番号の形式が不正です（例: 090-1234-5678）")
	}
	validShifts := map[string]bool{"morning": true, "afternoon": true, "night": true}
	if !validShifts[s.ShiftType] {
		return fmt.Errorf("シフト区分は morning/afternoon/night のいずれか: %s", s.ShiftType)
	}
	return nil
}

// Store はスタッフのインメモリストア。
type Store struct {
	mu     sync.RWMutex
	staffs map[string]*Staff
	nextID int
}

func NewStore() *Store {
	return &Store{staffs: make(map[string]*Staff), nextID: 1}
}

func (store *Store) Save(_ context.Context, s *Staff) error {
	if err := s.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if s.ID == "" {
		s.ID = fmt.Sprintf("staff-%d", store.nextID)
		store.nextID++
	}
	store.staffs[s.ID] = s
	return nil
}

func (store *Store) FindByID(_ context.Context, id string) (*Staff, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	s, ok := store.staffs[id]
	if !ok {
		return nil, fmt.Errorf("staff not found: %s", id)
	}
	return s, nil
}

func (store *Store) FindAll(_ context.Context) ([]*Staff, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]*Staff, 0, len(store.staffs))
	for _, s := range store.staffs {
		result = append(result, s)
	}
	return result, nil
}

func (store *Store) Delete(_ context.Context, id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if _, ok := store.staffs[id]; !ok {
		return fmt.Errorf("staff not found: %s", id)
	}
	delete(store.staffs, id)
	return nil
}
