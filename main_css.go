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
        &:hover { background-color: var(--fg80); }
    }
  }
}
`;
