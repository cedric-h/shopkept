package main

import (
	cryptrand "crypto/rand"
    "math"
	"math/rand"
	"log"
	"net/http"

    "slices"
    "maps"
	"fmt"
)

func gaussianRandom(mean, stddev float64) float64 {
    u1 := rand.Float64()
    u2 := rand.Float64()

    // Avoid log(0)
	if u1 == 0 {
		u1 = math.SmallestNonzeroFloat64
	}

    z0 := math.Sqrt(-2 * math.Log(u1)) * math.Cos(2 * math.Pi * u2)
    return z0 * stddev + mean
}

type Rcx struct {
	html, css string
}

func (s *Session) StatusBar(rcx *Rcx) {
    rcx.html += `<div class="status-bar">`

    rcx.html += `<div class="side">`
    rcx.html += `<span>⚜️ 0</span>`;
    rcx.html += `<span class="subtle">/4k to descend to L2</span>`;
    rcx.html += `</div>`

    rcx.html += `<div class="side right">`
    rcx.html += `<span>7:00 AM ☀️ </span>`;
    rcx.html += `<span class="subtle">day 1</span>`;
    rcx.html += `</div>`

    rcx.html += `</div>`

    rcx.css += `
        .status-bar {
            display: flex;
            justify-content: space-between;
            margin-bottom: 1rem;
            .right { align-items: flex-end; }
            .side {
                display: flex;
                flex-direction: column;
                justify-content: space-between;
                .subtle {
                    color: gray;
                    font-size: 0.5rem;
                }
            }
        }
    `
}

func (s *Session) TabBar(rcx *Rcx) {
	titleForTab := func(tab SessionTab) string {
		switch tab {
		case SessionTab_Inventory:
			return "👜 inventory"
		case SessionTab_Brewing:
			return "🚰 brewing"
		}
		return "wtf"
	}

	rcx.html += `<div class="tab-bar">`
	for t := SessionTab(0); t < SessionTab_COUNT; t++ {
		if s.Tab == t {
			rcx.html += `<div class="tab">`
			rcx.html += titleForTab(t)
			rcx.html += `</div>`
			continue
		}
		rcx.html += fmt.Sprintf(`<a class="tab" disabled href="tab%d">`, t)
		rcx.html += titleForTab(t)
		rcx.html += `</a>`
	}
	rcx.html += `</div>`
	rcx.css += `
        .tab-bar {
            font-size: 0.8rem;
            margin-bottom: 1rem;
            .tab {
                border-top-left-radius: 0.3rem;
                border-top-right-radius: 0.3rem;
                padding: 0.2rem 0.4rem 0.2rem 0.4rem;
                border: 0.1rem solid var(--fg);
                &[disabled] {
                    border: 0.1rem solid var(--disabled-fg);
                    border-bottom-width: 0rem;
                }
                border-bottom-width: 0rem;
            }
            width: 100%;
            display: flex;
            justify-content: space-evenly;
            border-bottom: 0.1rem solid var(--disabled-fg);
        }
    `
}

type InventoryContentActionKind uint
const (
	InventoryContentActionKind_None = iota
    InventoryContentActionKind_Open
)
type InventoryContentAction struct {
    kind InventoryContentActionKind
    open struct {
        got []struct { item Item; count uint }
    }
}

func (s *Session) InventoryContent(rcx *Rcx, invAction InventoryContentAction) {
    rcx.html += `<div class="inv">`
	for _, item := range slices.Sorted(maps.Keys(s.Inv)) {
        count := s.Inv[item]
		rcx.html += fmt.Sprintf(`<a href="item%d" class="inv-entry">`, item)
        rcx.html += `<span>` + item.Emoji() + `</span> `
        rcx.html += fmt.Sprintf(`<b>x%d</b> `, count)
        rcx.html += `<span>` + item.Title() + `</span> `
		rcx.html += `</a>`
	}
    rcx.html += `</div>`

    if s.InvTab.SelectedItem != Item_None {
        item := s.InvTab.SelectedItem
        rcx.html += `<div`
        rcx.html += ` class="inv-selected-modal-wrapper"`
        rcx.html += ` onclick="location.pathname='/item0'"`
        rcx.html += `>`
        rcx.html += `<div onclick="event.stopPropagation()" class="inv-selected-modal">`

        rcx.html += fmt.Sprintf(`<div class="title">%s</div>`, item.Title())
        rcx.html += fmt.Sprintf(`<div class="icon">%s</div>`, item.Emoji())
        rcx.html += fmt.Sprintf(`<i class="flavor">%s</i>`, item.Flavor())

        href, action := item.Action()
        if invAction.kind == InventoryContentActionKind_Open {
            href = "/item0"
            action = "done"

            if len(invAction.open.got) == 0 {
                rcx.html += "you got nothing ..."
            }

            rcx.html += `<div class="opened-got">`
            for _, got := range invAction.open.got {
                if got.count == 0 {
                    continue
                }
                rcx.html += `<div class="opened-got-entry">`
                rcx.html += got.item.Emoji()
                rcx.html += ` `
                rcx.html += fmt.Sprintf(`<b>x%d</b>`, got.count)
                rcx.html += ` `
                rcx.html += got.item.Title()
                rcx.html += `</div>`
            }
            rcx.html += `</div>`

            rcx.css += `
                .opened-got {
                    display: flex;
                    flex-direction: column;
                    align-items: flex-start;
                }
            `
        }
        rcx.html += fmt.Sprintf(
            `<a class="action" href="%s">%s</a>`,
            href,
            action,
        )
        rcx.html += `</div>`
        rcx.html += `</div>`

        rcx.css += `
            .inv-selected-modal-wrapper {
                background-color: var(--bg30);
                position: absolute;
                top: calc(-1 * var(--main-content-pad));
                bottom: calc(-1 * var(--main-content-pad));
                right: calc(-1 * var(--main-content-pad));
                left: calc(-1 * var(--main-content-pad));
                display: flex;
                align-items: center;
                justify-content: center;
            }
            .inv-selected-modal {
                display: flex;
                flex-direction: column;
                align-items: center;
                text-align: center;
                padding: 1.0rem 0.2rem 0.9rem 0.2rem;
                gap: 0.6rem;

                width: calc(90% - 2*var(--main-content-pad));
                border: 0.1rem solid var(--fg);
                border-radius: 0.3rem;
                background-color: var(--bg);

                .title {
                    font-size: 1rem;
                    font-weight: bold;
                }

                .icon {
                    font-size: 4rem;
                }

                .flavor {
                    padding: 0.0rem 0.8rem 0.0rem 0.8rem;
                    margin-bottom: 0.4rem;
                    font-size: 0.5rem;
                }

                .action {
                    width: 80%;
                    padding: 0.5rem;

                    border: 0.1rem solid var(--fg);
                    border-radius: 0.2rem;
                }
            }
        `
    }

    rcx.css += `
        .inv {
            display: flex;
            flex-direction: column;
            gap: 0.9rem;
            .inv-entry {
                padding: 0.4rem 0.8rem 0.4rem 0.8rem;
                border: 0.1rem solid var(--fg);
                border-radius: 0.3rem;
            }
        }
    `
}

type Session struct {
	Inv map[Item]uint

	Tab SessionTab
    InvTab struct {
        SelectedItem Item
    }
}

func NewSession() *Session {
	return &Session{
		Tab: SessionTab_Inventory,
		Inv: map[Item]uint{
			Item_MonsterCrate: 5,
			Item_FlyAgaric:    1,
			Item_Bone:         2,
		},
	}
}

type SessionTab uint

const (
	SessionTab_Inventory = iota
	SessionTab_Brewing
	SessionTab_COUNT
)

type Item uint

const (
	Item_None = iota
	Item_Bone
	Item_FlyAgaric
	Item_MonsterCrate
	Item_AncientCrate
	Item_HealthPotion
	Item_SkeletonKey
)

func (item Item) Emoji() string {
	switch item {
	case Item_None:
		return "🚫"
	case Item_Bone:
		return "🦴"
	case Item_FlyAgaric:
		return "🍄"
	case Item_MonsterCrate:
		return "📦"
	case Item_AncientCrate:
		return "🔒"
	case Item_HealthPotion:
		return "🧋"
	case Item_SkeletonKey:
		return "🗝️"
	}
	return "wtf"
}

func (item Item) Title() string {
	switch item {
	case Item_None:
		return "N/A"
	case Item_Bone:
		return "Bone"
	case Item_FlyAgaric:
		return "Fly Agaric"
	case Item_MonsterCrate:
		return "Monster Crate"
	case Item_AncientCrate:
		return "Ancient Crate"
	case Item_HealthPotion:
		return "Health Boba"
	case Item_SkeletonKey:
		return "Skeleton Key"
	}
	return "wtf"
}

func (item Item) Action() (string, string) {
	switch item {
	case Item_FlyAgaric, Item_Bone:
		return fmt.Sprintf("/tab%d", SessionTab_Brewing), "go to brewing"
	case Item_MonsterCrate, Item_AncientCrate:
		return fmt.Sprintf("/openitem%d", item), "open"
	}
	return "/item0", "done"
}

func (item Item) Flavor() string {
	switch item {
	case Item_None:
		return "you probably shouldn't be able to see this."

	case Item_Bone:
		return "once part of a spooky scary skeleton!" + 
		       " try brewing it into a skeleton key."

	case Item_FlyAgaric:
		return "a freaky shroom! If you eat it raw you'll trip balls and die," +
            " but if you know what you're doing you can brew it into a potion."

	case Item_MonsterCrate:
		return "a mystery box of gross things some adventurer has collected" +
            " from monsters!"

	case Item_AncientCrate:
		return "a spooky box! could contain cool stuff. " +
            "opened with skeleton keys!"

	case Item_HealthPotion:
		return "a healthy concoction made from mushrooms. adventurers love " +
            "this stuff!"

	case Item_SkeletonKey:
		return "a spooky key. can be used to open ancient crates."

	}
	return "wtf"
}

func (item Item) String() string {
    return fmt.Sprintf("%q", item.Title())
}

func (s *Session) GiveItems(item Item, count uint) {
    if _, has := s.Inv[item]; has {
        s.Inv[item] += count
        return
    }
    s.Inv[item] = count
}

func (s *Session) TakeItems(item Item, takeCount uint) bool {
    count, has := s.Inv[item];
    if has == false || count < takeCount {
        return false
    }

    s.Inv[item] -= takeCount

    if s.Inv[item] == 0 {
        delete(s.Inv, item)
    }

    return true
}

type Handler struct {
	sessions map[string]*Session
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	var sesh *Session
	{
		seshCookie, err := r.Cookie("Sesh")
		seshCookieVal := ""
		resetCookie := false
		if err != nil {
			resetCookie = true
			seshCookieVal = cryptrand.Text()
		} else {
			seshCookieVal = seshCookie.Value
		}

		if _, has := h.sessions[seshCookieVal]; !has {
			resetCookie = true
			h.sessions[seshCookieVal] = NewSession()
		}

		sesh = h.sessions[seshCookieVal]

		if resetCookie {
			/* ignore path bc we didn't know who they were */
			path = "/"

			/* erase their last action from the url bar */
			rw.Header().Set("Location", "/")
			http.SetCookie(rw, &http.Cookie{
				Name:     "Sesh",
				Value:    seshCookieVal,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})
			rw.WriteHeader(302)
		}
	}

	rcx := &Rcx{
		css: main_css,
	}

    sesh.StatusBar(rcx)

	{
		var tab SessionTab = 0
		n, err := fmt.Sscanf(path, "/tab%d", &tab)
		if n > 0 && err == nil {
			sesh.Tab = tab
		}

		sesh.TabBar(rcx)
	}

	switch sesh.Tab {

	case SessionTab_Inventory:
		action := InventoryContentAction{}

		var item Item = 0
		n, err := fmt.Sscanf(path, "/item%d", &item)
		if n > 0 && err == nil {
			sesh.InvTab.SelectedItem = item
		}
		n, err = fmt.Sscanf(path, "/openitem%d", &item)
        if n > 0 && err == nil {
            if item == Item_MonsterCrate &&
                sesh.TakeItems(Item_MonsterCrate, 1) {
                action.kind = InventoryContentActionKind_Open
                gaussianInt := func (mean, sttdev float64) uint {
                    return uint(max(0, math.Round(gaussianRandom(mean, sttdev))))
                }

                action.open.got = []struct { item Item; count uint }{
                    { Item_FlyAgaric, gaussianInt(2, 2) },
                    { Item_Bone, gaussianInt(2, 1) },
                }

                for _, drop := range action.open.got {
                    sesh.GiveItems(drop.item, drop.count)
                }
            }
        }

		sesh.InventoryContent(rcx, action)

	case SessionTab_Brewing:
		rcx.html += "Brewing content ..."
	}

	doc := fmt.Sprintf(`
<!DOCTYPE html>
<html lang='en'>
  <head>
    <meta charset='utf-8'/>
    <link rel="icon" href="data:image/svg+xml,<svg xmlns=%%22http://www.w3.org/2000/svg%%22 viewBox=%%220 0 100 100%%22><text y=%%22.9em%%22 font-size=%%2290%%22>👜</text></svg>">
    <title>shopkept</title>
    <style>
%s
    </style>
  </head>

  <body>
    <main>
      <div class="main-content">
%s
      </div>
    </main>
  </body>
</html>
    `, rcx.css, rcx.html)

	rw.Write([]byte(doc))
}

func main() {
	srv := http.Server{
		Addr:    ":8085",
		Handler: Handler{sessions: map[string]*Session{}},
	}

	log.Printf("hosting server on :8085")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
