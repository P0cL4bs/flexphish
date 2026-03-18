/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,ts}'],
  theme: {
    extend: {},
    fontFamily: {
      inter: ['inter', 'sans-serif'],
      poppins: ['poppins', 'serif'],
      'fira-mono': ['fira-mono'],
    },
  },
  safelist: [
    'badge-success',
    'badge-error',
    'badge-info',
    'badge-warning',
    'badge-neutral',
    'badge-secondary',
    'badge-primary',
  ],
  plugins: [require('@tailwindcss/typography'), require("daisyui")],
}
