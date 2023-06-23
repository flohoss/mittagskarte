/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  theme: {
    container: {
      center: true
    },
    extend: {
      animation: {
        fadeIn: 'fadeIn 0.3s ease-in',
        fadeIn50: 'fadeIn50 0.3s ease-in',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: 0 },
          '100%': { opacity: 1 },
        },
        fadeIn50: {
          '0%': { opacity: 0 },
          '100%': { opacity: 0.5 },
        }
      }
    }
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: [
      {
        light: {
          ...require("daisyui/src/theming/themes")["[data-theme=light]"],
          primary: "#EB912D",
          "primary-content": "#000",
          secondary: "#14468c",
          "secondary-content": "#fff",
          warning: "#EB912D",
          "warning-content": "#000",
          info: "#14468c"
        },
      },
      {
        dark: {
          ...require("daisyui/src/theming/themes")["[data-theme=dark]"],
          primary: "#EB912D",
          "primary-content": "#000",
          secondary: "#14468c",
          "secondary-content": "#fff",
          warning: "#EB912D",
          "warning-content": "#000",
          info: "#1a5cb9"
        },
      },
    ],
  },
};
