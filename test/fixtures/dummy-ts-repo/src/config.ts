// NOTE: this file intentionally contains fake placeholder credentials so that
// the security scanner has something to flag during integration tests. The
// values are not real secrets and grant no access to anything.

export const config = {
  awsAccessKey: "AKIAIOSFODNN7EXAMPLE",
  githubToken: "ghp_abcdef1234567890abcdef1234567890abcdef",
  apiKey: "Xq8aF3kL9pZ2nVbT5hY1wRcM6uJ4dG7s",
  serviceUrl: "https://api.example.test/v1",
} as const;

export type AppConfig = typeof config;
