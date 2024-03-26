/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.html"],
  theme: {
    container: {
      center: true
    },
    extend: {
      animation: {
        fadeIn: 'fadeIn 0.3s ease-in',
      },
      keyframes: {
        fadeIn: {
          '0%': {
            opacity: 0,
            transform: 'translateY(1rem)'
          },
          '100%': {
            opacity: 1,
            transform: 'translateY(0)'
          },
        }
      }
    }
  },
  plugins: [require("daisyui")]
};
