package world

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"git.subcult.tv/subculture-collective/edda/internal/rules"
)

type setupDB struct {
	calls  int
	failAt int
	mode   string
	writes []string
}

var errSetupWrite = errors.New("injected setup write failure")

func (d *setupDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	d.calls++
	d.writes = append(d.writes, sql)
	if d.calls == d.failAt {
		return pgconn.CommandTag{}, errSetupWrite
	}
	if strings.Contains(sql, "UPDATE campaigns") {
		d.mode = args[0].(string)
	}
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}
func (d *setupDB) Query(context.Context, string, ...any) (pgx.Rows, error) { panic("unexpected Query") }
func (d *setupDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	_, err := d.Exec(ctx, sql, args...)
	return setupRow{mode: d.mode, err: err}
}

type setupRow struct {
	mode string
	err  error
}

func (r setupRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	*dest[0].(*string) = r.mode
	return nil
}

func TestStartupWorldSetupReportsEveryWriteFailure(t *testing.T) {
	// Clock, rules mode, every feat, then every skill write must propagate.
	writes := 2 + len(rules.DefaultFeats) + len(rules.DefaultSkills)
	for failAt := 1; failAt <= writes; failAt++ {
		t.Run(fmt.Sprintf("write %d", failAt), func(t *testing.T) {
			d := &setupDB{failAt: failAt}
			_, err := orchestrateStartupWorldSetup(context.Background(), "crunch", d, uuid.New())
			if !errors.Is(err, errSetupWrite) {
				t.Fatalf("got %v, want wrapped write failure", err)
			}
			if d.calls != failAt {
				t.Fatalf("executed %d writes after failure at %d", d.calls, failAt)
			}
		})
	}
}

func TestStartupWorldSetupModes(t *testing.T) {
	for _, requested := range []string{"", "narrative", "crunch"} {
		t.Run(requested, func(t *testing.T) {
			d := &setupDB{}
			got, err := orchestrateStartupWorldSetup(context.Background(), requested, d, uuid.New())
			if err != nil {
				t.Fatal(err)
			}
			want := requested
			if want == "" {
				want = "narrative"
			}
			if got != want || d.mode != want {
				t.Fatalf("got=%s saved=%s want=%s", got, d.mode, want)
			}
			wantCalls := 2
			if want == "crunch" {
				wantCalls += len(rules.DefaultFeats) + len(rules.DefaultSkills)
			}
			if d.calls != wantCalls {
				t.Fatalf("calls=%d want=%d", d.calls, wantCalls)
			}
		})
	}
}

func TestStartupWorldSetupRequiresDatabase(t *testing.T) {
	if _, err := orchestrateStartupWorldSetup(context.Background(), "narrative", nil, uuid.New()); err == nil {
		t.Fatal("expected unavailable setup error")
	}
}
