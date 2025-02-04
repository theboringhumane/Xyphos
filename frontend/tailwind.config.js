/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ['class'],
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    container: {
      center: true,
      padding: '2rem',
      screens: {
        '2xl': '1400px',
      },
    },
    extend: {
      colors: {
        // Glitch Brand Colors
        'glitch-blue': {
          '50': '#e6f3ff',
          '100': '#bde0ff',
          '200': '#94cdff',
          '300': '#6bb9ff',
          '400': '#42a6ff',
          '500': '#2196f3',
          '600': '#1976d2',
          '700': '#1565c0',
          '800': '#0d47a1',
          '900': '#0a2472',
        },
        'glitch-purple': {
          '50': '#f3e5f5',
          '100': '#e1bee7',
          '200': '#ce93d8',
          '300': '#ba68c8',
          '400': '#ab47bc',
          '500': '#9c27b0',
          '600': '#8e24aa',
          '700': '#7b1fa2',
          '800': '#6a1b9a',
          '900': '#4a148c',
        },
        'glitch-pink': {
          '50': '#fce4ec',
          '100': '#f8bbd0',
          '200': '#f48fb1',
          '300': '#f06292',
          '400': '#ec407a',
          '500': '#e91e63',
          '600': '#d81b60',
          '700': '#c2185b',
          '800': '#ad1457',
          '900': '#880e4f',
        },
        'glitch-green': {
          '50': '#e8f5e9',
          '100': '#c8e6c9',
          '200': '#a5d6a7',
          '300': '#81c784',
          '400': '#66bb6a',
          '500': '#4caf50',
          '600': '#43a047',
          '700': '#388e3c',
          '800': '#2e7d32',
          '900': '#1b5e20',
        },
        'glitch-yellow': {
          '50': '#fffde7',
          '100': '#fff9c4',
          '200': '#fff59d',
          '300': '#fff176',
          '400': '#ffee58',
          '500': '#ffeb3b',
          '600': '#fdd835',
          '700': '#fbc02d',
          '800': '#f9a825',
          '900': '#f57f17',
        },
        'glitch-red': {
          '50': '#ffebee',
          '100': '#ffcdd2',
          '200': '#ef9a9a',
          '300': '#e57373',
          '400': '#ef5350',
          '500': '#f44336',
          '600': '#e53935',
          '700': '#d32f2f',
          '800': '#c62828',
          '900': '#b71c1c',
        },
        'glitch-gray': {
          '50': '#fafafa',
          '100': '#f5f5f5',
          '200': '#eeeeee',
          '300': '#e0e0e0',
          '400': '#bdbdbd',
          '500': '#9e9e9e',
          '600': '#757575',
          '700': '#616161',
          '800': '#424242',
          '900': '#212121',
        },
        // Semantic Colors
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'var(--glitch-blue-500)',
          foreground: 'hsl(var(--primary-foreground))',
          ...Array.from({ length: 9 }, (_, i) => i + 1).reduce((acc, i) => ({
            ...acc,
            [i * 100]: `var(--glitch-blue-${i * 100})`,
          }), {}),
        },
        secondary: {
          DEFAULT: 'var(--glitch-purple-500)',
          foreground: 'hsl(var(--secondary-foreground))',
          ...Array.from({ length: 9 }, (_, i) => i + 1).reduce((acc, i) => ({
            ...acc,
            [i * 100]: `var(--glitch-purple-${i * 100})`,
          }), {}),
        },
        success: {
          DEFAULT: 'var(--glitch-green-500)',
          foreground: 'hsl(var(--success-foreground))',
        },
        error: {
          DEFAULT: 'var(--glitch-red-500)',
          foreground: 'hsl(var(--error-foreground))',
        },
        warning: {
          DEFAULT: 'var(--glitch-yellow-500)',
          foreground: 'white',
        },
        muted: {
          DEFAULT: 'var(--glitch-gray-500)',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        border: 'var(--glitch-gray-300)',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
      fontFamily: {
        glitch: ['var(--font-family)', 'system-ui', 'sans-serif'],
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'slide-in': {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        'scale-in': {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
      animation: {
        'fade-in': 'fade-in 0.3s ease-in-out',
        'slide-in': 'slide-in 0.3s ease-in-out',
        'scale-in': 'scale-in 0.3s ease-in-out',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
}

