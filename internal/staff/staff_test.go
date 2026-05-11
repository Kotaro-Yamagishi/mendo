package staff_test

import (
	"context"
	"strings"
	"testing"

	"mendo/internal/staff"
)

// =============================================================================
// Validate
// =============================================================================

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		s             staff.Staff
		wantErr       bool
		wantErrContain string
	}{
		{
			name:    "正常系_morning",
			s:       staff.Staff{Name: "山田 太郎", Phone: "090-1234-5678", ShiftType: "morning"},
			wantErr: false,
		},
		{
			name:    "正常系_afternoon",
			s:       staff.Staff{Name: "佐藤", Phone: "080-9876-5432", ShiftType: "afternoon"},
			wantErr: false,
		},
		{
			name:    "正常系_night",
			s:       staff.Staff{Name: "佐藤", Phone: "080-9876-5432", ShiftType: "night"},
			wantErr: false,
		},
		{
			name:           "名前が空白のみ",
			s:              staff.Staff{Name: "   ", Phone: "090-1234-5678", ShiftType: "afternoon"},
			wantErr:        true,
			wantErrContain: "名前",
		},
		{
			name:           "電話番号_ハイフンなし",
			s:              staff.Staff{Name: "田中", Phone: "09012345678", ShiftType: "morning"},
			wantErr:        true,
			wantErrContain: "電話番号",
		},
		{
			name:    "電話番号_短すぎる",
			s:       staff.Staff{Name: "田中", Phone: "090-12", ShiftType: "morning"},
			wantErr: true,
		},
		{
			name:           "シフト区分_不正値",
			s:              staff.Staff{Name: "田中", Phone: "090-1234-5678", ShiftType: "evening"},
			wantErr:        true,
			wantErrContain: "シフト区分",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.s.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErrContain)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// =============================================================================
// Store.Save
// =============================================================================

func TestStore_Save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		input         staff.Staff
		wantErr       bool
		wantIDPrefix  string
		wantIDNotEmpty bool
	}{
		{
			name:           "IDが自動採番される",
			input:          staff.Staff{Name: "鈴木", Phone: "070-1111-2222", ShiftType: "night"},
			wantErr:        false,
			wantIDNotEmpty: true,
			wantIDPrefix:   "staff-",
		},
		{
			name:    "バリデーションエラー_名前空",
			input:   staff.Staff{Name: "", Phone: "090-1234-5678", ShiftType: "morning"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := staff.NewStore()
			s := tt.input
			err := store.Save(context.Background(), &s)
			if tt.wantErr {
				if err == nil {
					t.Error("expected validation error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Save failed: %v", err)
			}
			if tt.wantIDNotEmpty && s.ID == "" {
				t.Error("ID should be auto-assigned")
			}
			if tt.wantIDPrefix != "" && !strings.HasPrefix(s.ID, tt.wantIDPrefix) {
				t.Errorf("ID format unexpected: %q", s.ID)
			}
		})
	}
}

func TestStore_Save_IDIncrement(t *testing.T) {
	t.Parallel()
	store := staff.NewStore()
	ctx := context.Background()

	s1 := &staff.Staff{Name: "A", Phone: "090-0000-0001", ShiftType: "morning"}
	s2 := &staff.Staff{Name: "B", Phone: "090-0000-0002", ShiftType: "afternoon"}

	if err := store.Save(ctx, s1); err != nil {
		t.Fatalf("Save s1 failed: %v", err)
	}
	if err := store.Save(ctx, s2); err != nil {
		t.Fatalf("Save s2 failed: %v", err)
	}

	if s1.ID == s2.ID {
		t.Errorf("IDs should differ: %q == %q", s1.ID, s2.ID)
	}
}

// =============================================================================
// Store.FindByID
// =============================================================================

func TestStore_FindByID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		setup          func(store *staff.Store) string // 検索に使う ID を返す
		wantErr        bool
		wantErrContain string
		wantName       string
	}{
		{
			name: "存在するID",
			setup: func(store *staff.Store) string {
				s := &staff.Staff{Name: "高橋", Phone: "090-2222-3333", ShiftType: "afternoon"}
				if err := store.Save(context.Background(), s); err != nil {
					t.Fatalf("Save failed: %v", err)
				}
				return s.ID
			},
			wantErr:  false,
			wantName: "高橋",
		},
		{
			name:           "存在しないID",
			setup:          func(_ *staff.Store) string { return "staff-999" },
			wantErr:        true,
			wantErrContain: "staff not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := staff.NewStore()
			id := tt.setup(store)

			found, err := store.FindByID(context.Background(), id)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("error message unexpected: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("FindByID failed: %v", err)
			}
			if found.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", found.Name, tt.wantName)
			}
		})
	}
}

// =============================================================================
// Store.FindAll
// =============================================================================

func TestStore_FindAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		seeds     []staff.Staff
		wantCount int
	}{
		{
			name:      "空のストア",
			seeds:     nil,
			wantCount: 0,
		},
		{
			name: "2件保存済み",
			seeds: []staff.Staff{
				{Name: "X", Phone: "090-1111-0001", ShiftType: "morning"},
				{Name: "Y", Phone: "090-1111-0002", ShiftType: "night"},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := staff.NewStore()
			ctx := context.Background()

			for i := range tt.seeds {
				if err := store.Save(ctx, &tt.seeds[i]); err != nil {
					t.Fatalf("Save seed[%d]: %v", i, err)
				}
			}

			all, err := store.FindAll(ctx)
			if err != nil {
				t.Fatalf("FindAll failed: %v", err)
			}
			if len(all) != tt.wantCount {
				t.Errorf("FindAll returned %d, want %d", len(all), tt.wantCount)
			}
		})
	}
}

// =============================================================================
// Store.Delete
// =============================================================================

func TestStore_Delete(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		setup          func(store *staff.Store) string // 削除対象 ID を返す
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "存在するID",
			setup: func(store *staff.Store) string {
				s := &staff.Staff{Name: "伊藤", Phone: "080-4444-5555", ShiftType: "night"}
				if err := store.Save(context.Background(), s); err != nil {
					t.Fatalf("Save failed: %v", err)
				}
				return s.ID
			},
			wantErr: false,
		},
		{
			name:           "存在しないID",
			setup:          func(_ *staff.Store) string { return "staff-nonexistent" },
			wantErr:        true,
			wantErrContain: "staff not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := staff.NewStore()
			ctx := context.Background()
			id := tt.setup(store)

			err := store.Delete(ctx, id)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error for missing ID, got nil")
				}
				if tt.wantErrContain != "" && !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("error message unexpected: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}

			_, err = store.FindByID(ctx, id)
			if err == nil {
				t.Error("expected not found after delete")
			}
		})
	}
}
