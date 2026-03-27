// vim: noexpandtab
package main

import (
	cryptrand "crypto/rand"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"fmt"
	"maps"
	"slices"
)

func gaussianRandom(mean, stddev float64) float64 {
	u1 := rand.Float64()
	u2 := rand.Float64()

	/* avoid log(0) */
	if u1 == 0 {
		u1 = math.SmallestNonzeroFloat64
	}

	z0 := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return z0*stddev + mean
}

type Rcx struct {
	html, css, js  string
	refreshSeconds uint
}

func (rcx *Rcx) refreshWhen(when time.Time) {
	rcx.refreshSeconds = min(
		rcx.refreshSeconds,
		uint(math.Ceil(float64(when.Sub(time.Now()))/float64(time.Second))),
	)
}

func (s *Session) StatusBar(rcx *Rcx) {
	rcx.html += `<div class="status-bar">`

	rcx.html += `<div class="side">`
	rcx.html += fmt.Sprintf(
		`<span>⚜️ %d</span>`,
		s.Fleurs,
	)
	rcx.html += `<span class="subtle">/4k to descend to L2</span>`
	rcx.html += `</div>`

	rcx.html += `<div class="side right">`
	rcx.html += fmt.Sprintf(
		`<time data-format="gameclock" datetime="%s"></span>`,
		s.DayStart.Format(time.RFC3339),
	)
	rcx.html += `<span class="subtle">day 1</span>`
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

func (tradeable Tradeable) Emoji() string {
	switch tradeable.Kind {
		case TradeableKind_Item:
			return tradeable.Item.Emoji()
		case TradeableKind_Money:
			return "⚜️"
	}
	return "wtf"
}

func (tradeable Tradeable) Flavor() string {
	switch tradeable.Kind {
		case TradeableKind_Item:
			return tradeable.Item.Flavor()
		case TradeableKind_Money:
			return `golden fleurs! useful for buying things or ` +
				`descending deeper into the dungeon`
	}
	return "wtf"
}


func (s *Session) TradeOfferModal(rcx *Rcx) {
	trade := s.Trades[0]

	rcx.html += `<div`
	rcx.html += ` class="modal-wrapper"`
	rcx.html += ` onclick="location.pathname='/tradeaction0'"`
	rcx.html += `>`
	rcx.html += `<div onclick="event.stopPropagation()" class="modal trade-modal">`

	rcx.html += `<div class="trade-top">`
	{
		rcx.html += `<div class="trade-details">`

		rcx.html += `<div class="offer theirs">`
		rcx.html += `<div>YOU GIVE</div>`
		for _, t := range trade.YouGive {
			rcx.html += fmt.Sprintf(
				`<div>- %s x%d </div>`,
				t.Emoji(),
				t.Quantity,
			)
		}
		rcx.html += `</div>`

		rcx.html += `<div class="offer yours">`
		rcx.html += `<div>YOU TAKE</div>`
		for _, t := range trade.YouTake {
			rcx.html += fmt.Sprintf(
				`<div>+ %s x%d </div>`,
				t.Emoji(),
				t.Quantity,
			)
		}
		rcx.html += `</div>`

		rcx.html += `</div>`

		rcx.css += `
			.trade-details {
				display: flex;
				flex-direction: column;
				justify-content: center;
				gap: 1rem;

				.offer {
					display: flex;
					flex-direction: column;
					align-items: center;

					&.theirs { color: hsl(0, 100%, 80%); }
					&.yours { color: hsl(90, 100%, 80%); }
				}
			}
		`
	}

	{
		rcx.html += `<div class="trade-right-column">`

		rcx.html += `<div class="clock">⏰</div>`
		rcx.html += `<div>🤺</div>`

		rcx.html += `</div>`

		secondsIn := time.Now().Sub(trade.StartsAt) / time.Second
		duration := trade.EndsAt.Sub(trade.StartsAt) / time.Second
		rcx.refreshWhen(trade.EndsAt)

		growAnimation := fmt.Sprintf(
			`%ds ease-out -%ds clock-grow forwards`,
			duration,
			secondsIn,
		)

		flashAnimation := fmt.Sprintf(
			`0.5s linear %ds infinite alternate clock-flash;`,
			trade.EndsAt.Add(-3 * time.Second).Sub(time.Now()) / time.Second,
		)

		rcx.css += `
			@keyframes clock-grow {
				from { clip-path: circle(20% at 50% -10%); opacity: 1; }
				to { clip-path: circle(60% at 50% 50%); opacity: 1; }
			}
			@keyframes clock-shake {
				from { transform: rotate(-15deg); }
				to { transform: rotate(15deg); }
			}
			@keyframes clock-flash {
				from { opacity: 1; }
				35% { opacity: 1; }
				50% { opacity: 0.5; }
				65% { opacity: 1; }
				to { opacity: 1; }
			}
			.trade-right-column {
				.clock {
					position: relative;
					animation:
						0.3s cubic-bezier(0.68, -0.55, 0.27, 1.55) infinite alternate clock-shake,
						` + flashAnimation + `
					&::after {
						filter: grayscale(1);
						animation: ` + growAnimation + `;
						position: absolute;
						top: 0;
						bottom: 0;
						left: 0;
						right: 0;
						content: '⏰';
					}
				}
				display: flex;
				flex-direction: column;
				font-size: 2.5rem;
				gap: 0.5rem;
			}
		`
	}
	rcx.html += `</div>` //  class="trade-top"

	if len(trade.YouTake) > 0 {
		rcx.html += `<i class="flavor">` + trade.YouTake[0].Flavor() + `</i>`
	}

	rcx.html += `<div class="trade-options">`
	rcx.html += `<a class="no" href="/tradeaction0">no</a>`
	attrs := `href="/tradeaction1"`
	for _, t := range trade.YouGive {
		if !s.HasTradeable(t) {
			attrs = "disabled"
			break
		}
	}
	rcx.html += fmt.Sprintf(
		`<a class="yes" %s>yes</a>`,
		attrs,
	)
	rcx.html += `</div>`

	rcx.html += `</div>`
	rcx.html += `</div>`

	rcx.css += `
		.trade-modal {
			display: flex;
			flex-direction: column;
			gap: 1rem;

			.trade-top {
				display: flex;
				flex-direction: row;
				justify-content: space-around;
				width: 100%;
			}

			.flavor {
				padding: 0.0rem 0.8rem 0.0rem 0.8rem;
				margin-bottom: 0.4rem;
				font-size: 0.5rem;
				color: gray;
				text-align: center;
			}

			.trade-options {
				display: flex;
				flex-direction: row;
				justify-content: space-around;
				padding-left: 0.4rem;
				padding-right: 0.4rem;
				gap: 0.6rem;

				a {
					flex: 1;
					text-align: center;
					padding: 0.2rem 0.6rem 0.2rem 0.6rem;
					border: 0.1rem solid var(--fg);
					border-radius: 0.3rem;
					&.yes:not([disabled]) {
						color: hsl(209deg 85.57% 61.96%);
						background-color: hsl(209deg 85.57% 61.96% / 20%);
						&:hover {
							background-color: hsl(209deg 85.57% 61.96% / 60%);
							color: black;
						}
					}
					&.no {
						color: hsl(0deg 85.57% 61.96%);
						background-color: hsla(0deg 85.57% 61.96% / 20%);
						&:hover {
							background-color: hsl(0deg 85.57% 61.96% / 60%);
							color: black;
						}
					}
				}
			}
		}
	`;

}

func (s *Session) BrewingContent(rcx *Rcx) {
	rcx.html += `<div class="brew">`

	for i, bru := range s.Bru {
		stage := bru.BruStage()

		var title, subtitle, icon string
		switch stage {
		case BruStage_Empty:
			title = "EMPTY"
			subtitle = "click here to bru"
			icon = "🥽"
		case BruStage_Brewing:
			title = fmt.Sprintf(
				`<time datetime="%s" data-format="duration"></time> to go`,
				bru.Done.Format(time.RFC3339),
			)
			subtitle = "brewing ..."
			icon = "🧪" + bru.Recipe.Out().Emoji()
		case BruStage_Brewed:
			icon = bru.Recipe.Out().Emoji()
			title = "DONE"
			subtitle = fmt.Sprintf(
				`done <time datetime="%s" data-format="duration"></time> ago`,
				bru.Done.Format(time.RFC3339),
			)
		}

		href := ""
		if stage == BruStage_Brewing {
			rcx.refreshWhen(bru.Done.Add(-100 * time.Millisecond))
		}
		if stage != BruStage_Brewing {
			href = fmt.Sprintf(`href="/brew%d"`, i)
		}

		rcx.html += fmt.Sprintf(
			`<a %s class="brew-entry">`,
			href,
		)
		rcx.html += fmt.Sprintf(
			`<div class="brew-icon">%s</div>`,
			icon,
		)
		rcx.html += `<div class="brew-label">`
		rcx.html += fmt.Sprintf(`<div>%s</div>`, title)
		rcx.html += fmt.Sprintf(`<div class="subtle">%s</div>`, subtitle)
		rcx.html += `</div>`
		rcx.html += `</a>`
	}

	rcx.html += fmt.Sprintf(
		`<a href="/brew%d" class="brew-entry">`,
		len(s.Bru),
	)
	rcx.html += `<div class="brew-icon">🚰</div>`
	rcx.html += `<div class="brew-label">`
	rcx.html += `<div>BUY ⚜️ 100</div>`
	rcx.html += `<div class="subtle">unlock more brewing stations!</div>`
	rcx.html += `</div>`
	rcx.html += `</a>`

	rcx.html += `</div>` /* class="brew" */

	if s.BruTab.Modal == BruModal_BrewWhat {
		rcx.html += `<div`
		rcx.html += ` class="modal-wrapper"`
		rcx.html += ` onclick="location.pathname='/brew-1'"`
		rcx.html += `>`
		rcx.html += `<div onclick="event.stopPropagation()" class="modal">`
		rcx.html += `<div>what's brewing?</div>`

		for recipe := BruRecipe(1); recipe < BruRecipe_COUNT; recipe++ {
			out := recipe.Out()
			in, inCount, inTime := recipe.In()

			disabled := ""
			if !s.HasItems(in, inCount) {
				disabled = "disabled"
			}

			rcx.html += fmt.Sprintf(
				`<a %s href="/brewrecipe%d" class="brew-opt">`,
				disabled,
				recipe,
			)
			rcx.html += fmt.Sprintf(
				`<div class="title"> <b>%s %s </b> </div>`,
				out.Emoji(),
				out.Title(),
			)
			rcx.html += `<div class="subtle"> ingredients: </div>`
			rcx.html += fmt.Sprintf(
				`<div %s class="ingredient"> <b>x%d</b> %s %s </div>`,
				disabled,
				inCount,
				in.Title(),
				in.Emoji(),
			)
			rcx.html += fmt.Sprintf(
				`<div class="ingredient"> takes %s ⏰ </div>`,
				inTime.String(),
			)
			rcx.html += `</a>`
		}

		rcx.html += `</div>`
		rcx.html += `</div>`

		rcx.css += `
            .modal {
                .brew-opt {
                    font-size: 0.8rem;
                    padding: 0.4rem;
                    gap: 0.25rem;
                    display: flex;
                    flex-direction: column;

                    border: 0.1rem solid var(--fg);
                    border-radius: 0.3rem;
                    width: 80%;

                    .title {
                        margin-bottom: 0.5rem;
                    }
                    .subtle {
                        color: gray;
                    }
                    .ingredient {
                        margin-left: 0.5rem;
                        &[disabled] {
                            color: red;
                        }
                    }
                }
                display: flex;
                gap: 1rem;
                flex-direction: column;
                align-items: center;
            }
        `
	}

	if s.BruTab.Modal == BruModal_Done {
		rcx.html += `<div`
		rcx.html += ` class="modal-wrapper"`
		rcx.html += ` onclick="location.pathname='/brew-1'"`
		rcx.html += `>`
		rcx.html += `<div onclick="event.stopPropagation()" class="modal">`

		item := s.Bru[s.BruTab.SelectedBruIdx].Recipe.Out()

		rcx.html += `<div>u bru'd u a</div>`
		rcx.html += fmt.Sprintf(`<div class="title">%s</div>`, item.Title())
		rcx.html += fmt.Sprintf(`<div class="icon">%s</div>`, item.Emoji())
		rcx.html += fmt.Sprintf(`<i class="flavor">%s</i>`, item.Flavor())
		rcx.html += `<a class="action" href="/brew-1">yay</a>`

		rcx.html += `</div>`
		rcx.html += `</div>`

		rcx.css += `
            .modal {
                display: flex;
                gap: 0.6rem;
                flex-direction: column;
                align-items: center;
                text-align: center;

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
        .brew {
            display: flex;
            flex-direction: column;
            gap: 0.8rem;
            .brew-entry {
                border: 0.1rem solid var(--fg);
                border-radius: 0.3rem;
                display: flex;
				padding: 0.4rem;

				.brew-icon {
					width: 25%;
					font-size: 1.4rem;
					display: flex;
					align-items: center;
					justify-content: center;
				}

                .brew-label {
					width: 80%;
					display: flex;
					align-items: center;
					justify-content: center;
					flex-direction: column;
                    .subtle {
                        font-size: 0.5rem;
                        color: gray;
                    }
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
		rcx.html += fmt.Sprintf(`<a class="tab not-selected" href="tab%d">`, t)
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
                &.not-selected {
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
	InventoryContentActionKind_None InventoryContentActionKind = iota
	InventoryContentActionKind_Open
)

type InventoryContentAction struct {
	kind InventoryContentActionKind
	open struct {
		got []struct {
			item  Item
			count uint
		}
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
		rcx.html += ` class="modal-wrapper"`
		rcx.html += ` onclick="location.pathname='/item0'"`
		rcx.html += `>`
		rcx.html += `<div onclick="event.stopPropagation()" class="modal inv-selected-modal">`

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
            .inv-selected-modal {
                display: flex;
                gap: 0.6rem;
                flex-direction: column;
                align-items: center;
                text-align: center;

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

type Bru struct {
	Recipe BruRecipe
	Done   time.Time
}

type BruRecipe uint

const (
	BruRecipe_NONE BruRecipe = iota
	BruRecipe_HealthBoba
	BruRecipe_SkeletonKey
	BruRecipe_COUNT
)

func (recipe BruRecipe) Out() Item {
	switch recipe {
	case BruRecipe_HealthBoba:
		return Item_HealthBoba
	case BruRecipe_SkeletonKey:
		return Item_SkeletonKey
	case BruRecipe_COUNT:
		return Item_None
	}
	return Item_None
}

func (recipe BruRecipe) In() (Item, uint, time.Duration) {
	switch recipe {
	case BruRecipe_HealthBoba:
		return Item_FlyAgaric, 5, 30 * time.Second
	case BruRecipe_SkeletonKey:
		return Item_Bone, 25, time.Minute
	case BruRecipe_COUNT, BruRecipe_NONE:
		return Item_None, 0, time.Second
	}
	return Item_None, 0, time.Second
}

type BruStage uint

const (
	BruStage_Brewing BruStage = iota
	BruStage_Brewed
	BruStage_Empty
)

type BruModal uint

const (
	BruModal_None BruModal = iota
	BruModal_BrewWhat
	BruModal_Done
)

func (bru Bru) BruStage() BruStage {
	if bru.Recipe == BruRecipe_NONE {
		return BruStage_Empty
	} else if bru.Done.Before(time.Now()) {
		return BruStage_Brewed
	} else {
		return BruStage_Brewing
	}
}

type TradeableKind uint
const (
	TradeableKind_Money TradeableKind = iota
	TradeableKind_Item 
)
type Tradeable struct {
	Kind TradeableKind
	Item Item
	Quantity uint
}

type Trade struct {
	EndsAt, StartsAt time.Time
	YouGive []Tradeable
	YouTake []Tradeable
}

type Session struct {
	DayStart time.Time
	Fleurs   uint
	Trades []Trade

	Inv map[Item]uint
	Bru []Bru

	Tab    SessionTab
	InvTab struct {
		SelectedItem Item
	}
	BruTab struct {
		Modal          BruModal
		SelectedBruIdx int
	}
}

func NewSession() *Session {
	return &Session{
		DayStart: time.Now(),
		Fleurs:   100,
		Tab:      SessionTab_Inventory,

		Inv: map[Item]uint{
			Item_MonsterCrate: 5,
			Item_FlyAgaric:    1,
			Item_Bone:         2,
			Item_HealthBoba:   1,
		},

		Bru: []Bru{
			{},
		},

		Trades: []Trade{
			{
				EndsAt: time.Now().Add(time.Second * 10),
				StartsAt: time.Now(),
				YouGive: []Tradeable {
					{
						Kind: TradeableKind_Money,
						Quantity: 5,
					},
				},
				YouTake: []Tradeable {
					{
						Kind: TradeableKind_Item,
						Item: Item_MonsterCrate,
						Quantity: 1,
					},
				},
			},
			{
				EndsAt: time.Now().Add(time.Second * 60),
				StartsAt: time.Now().Add(time.Second * 50),
				YouGive: []Tradeable {
					{
						Kind: TradeableKind_Item,
						Item: Item_HealthBoba,
						Quantity: 1,
					},
				},
				YouTake: []Tradeable {
					{
						Kind: TradeableKind_Money,
						Quantity: 15,
					},
				},
			},
		},
	}
}

type SessionTab uint

const (
	SessionTab_Inventory SessionTab = iota
	SessionTab_Brewing
	SessionTab_COUNT
)

type Item uint

const (
	Item_None Item = iota
	Item_Bone
	Item_FlyAgaric
	Item_MonsterCrate
	Item_AncientCrate
	Item_HealthBoba
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
	case Item_HealthBoba:
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
	case Item_HealthBoba:
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

	case Item_HealthBoba:
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

func (s *Session) TakeFleurs(count uint) bool {
	if s.Fleurs < count {
		return false
	}
	s.Fleurs -= count
	return true
}

func (s *Session) HasItems(item Item, takeCount uint) bool {
	count, has := s.Inv[item]
	return has && count >= takeCount
}

func (s *Session) TakeItems(item Item, takeCount uint) bool {
	if !s.HasItems(item, takeCount) {
		return false
	}

	s.Inv[item] -= takeCount

	if s.Inv[item] == 0 {
		delete(s.Inv, item)
	}

	return true
}

func (sesh *Session) HasTradeable(t Tradeable) bool {
	switch t.Kind {
		case TradeableKind_Money:
			return sesh.Fleurs >= t.Quantity
		case TradeableKind_Item:
			return sesh.HasItems(t.Item, t.Quantity)
	}
	return false
}

func (sesh *Session) TakeTradeable(t Tradeable) {
	switch t.Kind {
		case TradeableKind_Money:
			sesh.TakeFleurs(t.Quantity)
		case TradeableKind_Item:
			sesh.TakeItems(t.Item, t.Quantity)
	}
}

func (sesh *Session) GiveTradeable(t Tradeable) {
	switch t.Kind {
		case TradeableKind_Money:
			sesh.Fleurs += t.Quantity
		case TradeableKind_Item:
			sesh.GiveItems(t.Item, t.Quantity)
	}
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
		css:            main_css,
		js:             main_js,
		refreshSeconds: 9999999,
	}

	/* TradeOfferModal controller */
	if len(sesh.Trades) > 0 {
		trade := sesh.Trades[0]
		if time.Now().After(trade.StartsAt) && time.Now().Before(trade.EndsAt) {
				
			if path == "/tradeaction0" {
				sesh.Trades = sesh.Trades[1:]
			} else if path == "/tradeaction1" {

				func () {
					for _, t := range trade.YouGive {
						if !sesh.HasTradeable(t) {
							return
						}
					}

					for _, t := range trade.YouGive {
						sesh.TakeTradeable(t)
					}

					for _, t := range trade.YouTake {
						sesh.GiveTradeable(t)
					}

					sesh.Trades = sesh.Trades[1:]
				}()
			}
		}
	}

	/* TabBar controller */
	var tab SessionTab = 0
	n, err := fmt.Sscanf(path, "/tab%d", &tab)
	if n > 0 && err == nil {
		sesh.Tab = tab
	}
	topBars := func() {
		sesh.StatusBar(rcx)

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
				gaussianInt := func(mean, sttdev float64) uint {
					return uint(max(0, math.Round(gaussianRandom(mean, sttdev))))
				}

				action.open.got = []struct {
					item  Item
					count uint
				}{
					{Item_FlyAgaric, gaussianInt(2, 2)},
					{Item_Bone, gaussianInt(2, 1)},
				}

				for _, drop := range action.open.got {
					sesh.GiveItems(drop.item, drop.count)
				}
			}
		}

		topBars()
		sesh.InventoryContent(rcx, action)

	case SessionTab_Brewing:

		func() {
			var brewIdx int = 0
			n, err := fmt.Sscanf(path, "/brew%d", &brewIdx)
			if n == 0 || err != nil {
				return
			}

			if brewIdx < 0 {
				if sesh.BruTab.Modal == BruModal_Done {
					bru := sesh.Bru[sesh.BruTab.SelectedBruIdx]
					sesh.GiveItems(bru.Recipe.Out(), 1)
					sesh.Bru[sesh.BruTab.SelectedBruIdx] = Bru{}
				}

				sesh.BruTab.Modal = BruModal_None
				return
			}

			if brewIdx >= len(sesh.Bru) {
				if sesh.TakeFleurs(100) {
					sesh.Bru = append(sesh.Bru, Bru{})
				}
				return
			}

			bru := sesh.Bru[brewIdx]

			switch bru.BruStage() {
			case BruStage_Brewed:
				sesh.BruTab.Modal = BruModal_Done
				sesh.BruTab.SelectedBruIdx = brewIdx
			case BruStage_Brewing:
				/* no action */
			case BruStage_Empty:
				sesh.BruTab.Modal = BruModal_BrewWhat
				sesh.BruTab.SelectedBruIdx = brewIdx
			}
		}()

		func() {
			var recipeIdx int = 0
			n, err := fmt.Sscanf(path, "/brewrecipe%d", &recipeIdx)
			if n == 0 || err != nil {
				return
			}

			if recipeIdx >= int(BruRecipe_COUNT) {
				return
			}

			recipe := BruRecipe(recipeIdx)
			in, inCount, inTime := recipe.In()
			if sesh.TakeItems(in, inCount) {
				sesh.Bru[sesh.BruTab.SelectedBruIdx] = Bru{
					Recipe: recipe,
					Done:   time.Now().Add(inTime),
				}
				sesh.BruTab.Modal = BruModal_None
			}
		}()

		topBars()
		sesh.BrewingContent(rcx)
	}

	if len(sesh.Trades) > 0 {
		trade := sesh.Trades[0]
		if time.Now().After(trade.StartsAt) && time.Now().Before(trade.EndsAt) {
			sesh.TradeOfferModal(rcx)
		}
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
    <script>
%s
    </script>
  </head>

  <body>
    <main>
      <div class="main-content">
%s
      </div>
    </main>
  </body>
</html>
    `, rcx.css, rcx.js, rcx.html)

	rw.Header().Set("Refresh", fmt.Sprintf("%d", rcx.refreshSeconds))

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
