/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: "jit",
  content: ["./static/**/*.html"],
  theme: {
    extend: {
      fontFamily: {
        header: ["Noyh R"],
      },
      backgroundImage: {
        background: "url('/images/background-pattern.svg')",
      },
      boxShadow: {
        "custom-border": "0 4px 0 0 #181825",
      },
      colors: {
        rosewater: "#f5e0dc",
        flamingo: "#f2cdcd",
        pink: "#f5c2e7",
        mauve: "#cba6f7",
        red: "#f38ba8",
        maroon: "#eba0ac",
        peach: "#fab387",
        yellow: "#f9e2af",
        green: "#a6e3a1",
        teal: "#94e2d5",
        sky: "#89dceb",
        sapphire: "#74c7ec",
        blue: "#89b4fa",
        lavender: "#b4befe",
        text: "#cdd6f4",
        text2: "#D3D3E5",
        subtext1: "#bac2de",
        subtext0: "#a6adc8",
        overlay2: "#9399b2",
        overlay1: "#7f849c",
        overlay0: "#6c7086",
        surface2: "#585b70",
        surface1: "#45475a",
        surface0: "#313244",
        base: "#1e1e2e",
        mantle: "#181825",
        crust: "#11111b",
      },
    },
  },
  plugins: [],
};
