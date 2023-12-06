/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.html"],
  theme: {
    container: {
      padding: {
        DEFAULT: '0.5rem',
        lg: '2rem',
        xl: '12rem',
        '2xl': '18rem',
      },
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
          ...require("daisyui/src/theming/themes")["light"],
          "primary": "#EB912D",
          "primary-content": "#000",
          "secondary": "#14468c",
          "secondary-content": "#fff",
          "success": "#28a745",
          "success-content": "white",
          "error": "#dc3545",
          "error-content": "white",
          "warning": "#ffc107",
          "warning-content": "black",
          "info": "#17a2b8",
          "info-content": "white",
        },
      },
      {
        dark: {
          ...require("daisyui/src/theming/themes")["dark"],
          "primary": "#EB912D",
          "primary-content": "#000",
          "secondary": "#14468c",
          "secondary-content": "#fff",
          "success": "#28a745",
          "success-content": "white",
          "error": "#dc3545",
          "error-content": "white",
          "warning": "#ffc107",
          "warning-content": "black",
          "info": "#17a2b8",
          "info-content": "white",
        },
      },
    ],
  },
};
