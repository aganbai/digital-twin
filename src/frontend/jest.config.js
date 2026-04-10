/** @type {import('jest').Config} */
module.exports = {
  preset: 'ts-jest',
  testEnvironment: 'jsdom',
  roots: ['<rootDir>/src'],
  testMatch: ['**/__tests__/**/*.test.(ts|tsx)', '**/?(*.)+(spec|test).(ts|tsx)'],
  moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '^@tarojs/taro$': '<rootDir>/src/__mocks__/taro.ts',
    '^@tarojs/components$': '<rootDir>/src/__mocks__/components.ts',
    '\\.(scss|sass|css)$': '<rootDir>/src/__mocks__/styleMock.ts',
  },
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
  coverageDirectory: '<rootDir>/coverage',
  coverageReporters: ['text', 'text-summary', 'lcov', 'html'],
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/index.config.ts',
    '!src/**/*.scss',
    '!src/**/__mocks__/**',
    '!src/**/__tests__/**',
  ],
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx',
        esModuleInterop: true,
        allowSyntheticDefaultImports: true,
        noUnusedLocals: false,
        noUnusedParameters: false,
        skipLibCheck: true,
        strict: false,
        paths: {
          '@/*': ['./src/*'],
        },
        baseUrl: '.',
      },
      babelConfig: {
        presets: [
          ['@babel/preset-react', { runtime: 'automatic' }],
        ],
      },
      isolatedModules: true,
    }],
  },
  transformIgnorePatterns: [
    'node_modules/(?!(zustand|msw)/)',
  ],
  testPathIgnorePatterns: [
    '/node_modules/',
  ],
}
