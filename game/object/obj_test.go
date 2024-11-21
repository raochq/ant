package object

import (
	"context"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/raochq/ant/game/object/maps"
	"github.com/raochq/ant/game/object/npc"
	"github.com/raochq/ant/game/object/player"
)

func TestRun(t *testing.T) {
	// m := maps.NewStaticMap("static_map_1")
	m, _ := maps.NewActivityMap("Activity_Map_1", true)
	p, _ := player.NewPlayer(1, "hello", 10, true)
	n, _ := npc.NewNPC(2, "npc1", 5)
	b, _ := npc.NewBuilding(3, "building1")
	m.AddObject(p.Object())
	m.AddObject(n.Object())
	m.AddObject(b.Object())
	m.Tick(time.Now())
	for i := 0; i < 10; i++ {
		m.Tick(time.Now())
		time.Sleep(time.Second)
	}

	t.Log("done")
	// m.MoveObject(p, base.Point{X: 10, Y: 10})
	// m.MoveObject(n, base.Point{X: 10, Y: 10})
	// p.MoveTo(base.Point{X: 10, Y: 10})
	// p.MoveTo(base.Point{X: 10, Y: 1002})
	// m.Tick(time.Now())
}

type myHandler struct {
	slog.Handler
}

func (h *myHandler) Enabled(_ context.Context, lv slog.Level) bool {
	return lv >= slog.LevelDebug
}
func init() {
	oldFlag := log.Flags()
	slog.SetDefault(slog.New(&myHandler{
		slog.Default().Handler(),
	}))
	log.SetFlags(oldFlag)
	log.SetOutput(os.Stdout)
}
