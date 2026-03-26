package main

const main_css = `
document, body {
  width: 100vw; height: 100vh;
  margin: 0px; padding: 0px;
  font-family: monospace;
  --fg80: color-mix(in srgb, var(--fg), transparent 80%);
  --bg30: color-mix(in srgb, var(--bg), transparent 30%);
}
:root {
  color-scheme: light dark;
  font-size: calc(6 * min(1vw, 1vh * (9/16)));
  --disabled-fg: #888888;
  --main-content-pad: 1rem;
}
@media (prefers-color-scheme: dark) {
  :root {
    --fg: white;
    --bg: #121212;
  }
}
@media (prefers-color-scheme: light) {
  :root {
    --fg: black;
    --bg: white;
  }
}
main {
  aspect-ratio: 9/16;
  border: 0.1rem solid var(--fg);
  overflow: hidden;
  position: absolute;
  inset: 0;
  margin: auto;
  min-height: 0;
  max-height: calc(100% - 0.3rem);
  .main-content {
    position: relative;
    margin: var(--main-content-pad);
    width: calc(100% - 2*var(--main-content-pad));
    height: calc(100% - 2*var(--main-content-pad));

    a {
        text-decoration: none;
        cursor: pointer;
        &:hover:not([disabled]) { background-color: var(--fg80); }
        &[disabled] {
            color: gray;
            cursor: not-allowed;
        }
    }

    .modal-wrapper {
        .modal {
            padding: 1.0rem 0.2rem 0.9rem 0.2rem;

            width: calc(90% - 2*var(--main-content-pad));
            border: 0.1rem solid var(--fg);
            border-radius: 0.3rem;
            background-color: var(--bg);
        }
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
  }
}
`;

const main_js = `
function durationStr(timestamp) {
    const t = Math.abs(Date.now() - Date.parse(timestamp));

    const d = Math.floor(t / (24 * 60 * 60 * 1000));
    const h = Math.floor(t / (60 * 60 * 1000) % 24);
    const m = Math.floor(t / (60 * 1000) % 60);
    const s = Math.floor(t / 1000) % 60;

    if (s < 0.2) return 'now';
    if (m < 1) return s + 's';
    if (h < 1) return ` + "`${m}m ${s}s`" + `;
    if (d < 1) return ` + "`${h}h ${m}m ${s}s`" + `;
}

function gameclockStr(timestamp) {
    let t = Date.now() - Date.parse(timestamp);
    t /= 180;

    let h = Math.floor(t / 60) + 6;
    let m = Math.floor(t % 60);
    m = Math.floor(m/10)*10;

    /* time stops at 8:00 PM */
    if (h > 19) {
        h = 20
        m = 0
    }

    let emoji = '☀️';
    if (h < 7) emoji = '🌅';
    else if (h < 15) emoji = '☀️';
    else if (h < 21) emoji = '🌚';

    const am = (h < 12) ? 'AM' : 'PM';

    const mins = m.toString().padStart(2, '0');
    return ((h - 1) % 12 + 1) + ':' + mins + am + ' ' + emoji;
}

requestAnimationFrame(function frame() {
    requestAnimationFrame(frame);
    document
        .querySelectorAll("time[data-format]")
        .forEach(x => {
            if (x.dataset.format == "duration")
                x.textContent = durationStr(x.dateTime);
            if (x.dataset.format == "gameclock")
                x.textContent = gameclockStr(x.dateTime);
        })
})
`;
