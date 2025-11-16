import {heroui} from "@heroui/theme"

/** @type {import('tailwindcss').Config} */
const config = {
  content: [
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    "./node_modules/@heroui/theme/dist/**/*.{js,ts,jsx,tsx}"
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["var(--font-sans)"],
        mono: ["var(--font-mono)"],
      },
    },
  },
  darkMode: "class",
  plugins: [
    heroui({
  "themes": {
    "light": {
      "colors": {
        "default": {
          "50": "#fafafa",
          "100": "#f2f2f3",
          "200": "#ebebec",
          "300": "#e3e3e6",
          "400": "#dcdcdf",
          "500": "#d4d4d8",
          "600": "#afafb2",
          "700": "#8a8a8c",
          "800": "#656567",
          "900": "#404041",
          "foreground": "#000",
          "DEFAULT": "#d4d4d8"
        },
        "primary": {
          "50": "#fde5e4",
          "100": "#fac1bd",
          "200": "#f79d97",
          "300": "#f57971",
          "400": "#f2554a",
          "500": "#ef3124",
          "600": "#c5281e",
          "700": "#9b2017",
          "800": "#721711",
          "900": "#480f0b",
          "foreground": "#000",
          "DEFAULT": "#ef3124"
        },
        "secondary": {
          "50": "#dfeeff",
          "100": "#b3d6ff",
          "200": "#86beff",
          "300": "#59a7ff",
          "400": "#2d8fff",
          "500": "#0077ff",
          "600": "#0062d2",
          "700": "#004da6",
          "800": "#003979",
          "900": "#00244d",
          "foreground": "#000",
          "DEFAULT": "#0077ff"
        },
        "success": {
          "50": "#f4ffdf",
          "100": "#e5ffb3",
          "200": "#d6ff86",
          "300": "#c6ff59",
          "400": "#b7ff2d",
          "500": "#a8ff00",
          "600": "#8bd200",
          "700": "#6da600",
          "800": "#507900",
          "900": "#324d00",
          "foreground": "#000",
          "DEFAULT": "#a8ff00"
        },
        "warning": {
          "50": "#fffcdf",
          "100": "#fff7b3",
          "200": "#fff286",
          "300": "#ffed59",
          "400": "#ffe92d",
          "500": "#ffe400",
          "600": "#d2bc00",
          "700": "#a69400",
          "800": "#796c00",
          "900": "#4d4400",
          "foreground": "#000",
          "DEFAULT": "#ffe400"
        },
        "danger": {
          "50": "#ffe1df",
          "100": "#ffb7b3",
          "200": "#ff8d86",
          "300": "#ff6459",
          "400": "#ff3a2d",
          "500": "#ff1000",
          "600": "#d20d00",
          "700": "#a60a00",
          "800": "#790800",
          "900": "#4d0500",
          "foreground": "#000",
          "DEFAULT": "#ff1000"
        },
        "background": "#ffffff",
        "foreground": "#000000",
        "content1": {
          "DEFAULT": "#ffffff",
          "foreground": "#000"
        },
        "content2": {
          "DEFAULT": "#f4f4f5",
          "foreground": "#000"
        },
        "content3": {
          "DEFAULT": "#e4e4e7",
          "foreground": "#000"
        },
        "content4": {
          "DEFAULT": "#d4d4d8",
          "foreground": "#000"
        },
        "focus": "#006FEE",
        "overlay": "#000000"
      }
    },
    "dark": {
      "colors": {
        "default": {
          "50": "#1b1b1d",
          "100": "#363639",
          "200": "#515156",
          "300": "#6c6c72",
          "400": "#87878f",
          "500": "#9f9fa5",
          "600": "#b7b7bc",
          "700": "#cfcfd2",
          "800": "#e7e7e9",
          "900": "#ffffff",
          "foreground": "#000",
          "DEFAULT": "#87878f"
        },
        "primary": {
          "50": "#480f0b",
          "100": "#721711",
          "200": "#9b2017",
          "300": "#c5281e",
          "400": "#ef3124",
          "500": "#f2554a",
          "600": "#f57971",
          "700": "#f79d97",
          "800": "#fac1bd",
          "900": "#fde5e4",
          "foreground": "#000",
          "DEFAULT": "#ef3124"
        },
        "secondary": {
          "50": "#00244d",
          "100": "#003979",
          "200": "#004da6",
          "300": "#0062d2",
          "400": "#0077ff",
          "500": "#2d8fff",
          "600": "#59a7ff",
          "700": "#86beff",
          "800": "#b3d6ff",
          "900": "#dfeeff",
          "foreground": "#000",
          "DEFAULT": "#0077ff"
        },
        "success": {
          "50": "#324d00",
          "100": "#507900",
          "200": "#6da600",
          "300": "#8bd200",
          "400": "#a8ff00",
          "500": "#b7ff2d",
          "600": "#c6ff59",
          "700": "#d6ff86",
          "800": "#e5ffb3",
          "900": "#f4ffdf",
          "foreground": "#000",
          "DEFAULT": "#a8ff00"
        },
        "warning": {
          "50": "#4d4400",
          "100": "#796c00",
          "200": "#a69400",
          "300": "#d2bc00",
          "400": "#ffe400",
          "500": "#ffe92d",
          "600": "#ffed59",
          "700": "#fff286",
          "800": "#fff7b3",
          "900": "#fffcdf",
          "foreground": "#000",
          "DEFAULT": "#ffe400"
        },
        "danger": {
          "50": "#4d0500",
          "100": "#790800",
          "200": "#a60a00",
          "300": "#d20d00",
          "400": "#ff1000",
          "500": "#ff3a2d",
          "600": "#ff6459",
          "700": "#ff8d86",
          "800": "#ffb7b3",
          "900": "#ffe1df",
          "foreground": "#000",
          "DEFAULT": "#ff1000"
        },
        "background": "#000000",
        "foreground": "#ffffff",
        "content1": {
          "DEFAULT": "#18181b",
          "foreground": "#fff"
        },
        "content2": {
          "DEFAULT": "#27272a",
          "foreground": "#fff"
        },
        "content3": {
          "DEFAULT": "#3f3f46",
          "foreground": "#fff"
        },
        "content4": {
          "DEFAULT": "#52525b",
          "foreground": "#fff"
        },
        "focus": "#006FEE",
        "overlay": "#ffffff"
      }
    }
  },
  "layout": {
    "disabledOpacity": "0.5"
  }
}
  )],
}

module.exports = config;