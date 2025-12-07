/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "primary": "#7f13ec",
        "background-light": "#f7f6f8",
        "background-dark": "#191022",
      },
      fontFamily: {
        "display": ["Spline Sans", "sans-serif"],
        "sans": ["Spline Sans", "sans-serif"],
      },
      borderRadius: {
        "DEFAULT": "0.25rem",
        "lg": "0.5rem",
        "xl": "0.75rem",
        "full": "9999px"
      },
      backgroundImage: {
        'illustration-gradient': "radial-gradient(circle at 15% 50%, rgba(127, 19, 236, 0.2), transparent 40%), radial-gradient(circle at 85% 30%, rgba(127, 19, 236, 0.15), transparent 40%), radial-gradient(circle at 50% 90%, rgba(127, 19, 236, 0.1), transparent 50%)",
      }
    },
  },
  plugins: [],
}
