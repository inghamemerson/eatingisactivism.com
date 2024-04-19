/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./app/templates/**/*.{html,tmpl}"],
  theme: {
    extend: {
      colors: {
        "pg-purple": "#472987",
        "pg-blue": "#016BB7",
        "pg-orange": "#FF7B34",
      },
    },
    fontFamily: {
      sans: '"Avenir Next", system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", "Liberation Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"',
    },
  },
  plugins: [],
  safelist: [
    {
      pattern: /^location-popup/,
    },
    {
      pattern: /^button/,
    },
    {
      pattern: /^mapboxgl/,
    },
    "tag",
    "tags",
    "gold",
    "silver",
    "bronze",
    "patagonia",
  ],
};
