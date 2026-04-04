import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        obsidian: '#0A0A0A',
        charcoal: '#141414',
        champagne: '#F2F0E4',
        gold: {
          DEFAULT: '#D4AF37',
          light: '#F2E8C4',
          dark: '#B8901A',
        },
        midnight: '#1E3D59',
        pewter: '#888888',
        jade: {
          DEFAULT: '#4ECCA3',
          light: '#7EDBB8',
          dark: '#2BA47C',
        },
        ruby: {
          DEFAULT: '#AB2346',
          light: '#D14D6E',
          dark: '#8A1C38',
        },
        sapphire: '#5B8FB9',
      },
      fontFamily: {
        heading: ['Marcellus', 'serif'],
        sans: ['Josefin Sans', 'sans-serif'],
      },
      boxShadow: {
        'gold-sm': '0 0 10px rgba(212, 175, 55, 0.1)',
        'gold': '0 0 20px rgba(212, 175, 55, 0.15)',
        'gold-lg': '0 0 30px rgba(212, 175, 55, 0.25)',
        'jade-sm': '0 0 10px rgba(78, 204, 163, 0.1)',
        'jade': '0 0 20px rgba(78, 204, 163, 0.15)',
        'ruby-sm': '0 0 10px rgba(171, 35, 70, 0.12)',
        'ruby': '0 0 20px rgba(171, 35, 70, 0.18)',
        'sapphire-sm': '0 0 10px rgba(91, 143, 185, 0.12)',
        'sapphire': '0 0 20px rgba(91, 143, 185, 0.15)',
      },
    },
  },
} satisfies Config;
