export default {
  branches: ["main"],
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@semantic-release/exec",
      {
        verifyReleaseCmd: "echo ${nextRelease.version} > .VERSION",
      },
    ],
    [
      "@semantic-release/exec",
      {
        verifyReleaseCmd:
          "go run ./cmd/yeet --force-git-version=$(cat .VERSION)",
      },
    ],
    [
      "@semantic-release/github",
      {
        assets: ["var/**"],
      },
    ],
    [
      "@semantic-release/npm",
      {
        npmPublish: false,
      },
    ],
    [
      "@semantic-release/changelog",
      {
        changeLogFile: "CHANGLOG.md",
      },
    ],
    [
      "@semantic-release/git",
      {
        assets: ["CHANGELOG.md", "package.json"],
        message:
          "chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}\n\nSigned-Off-By: Mimi Yasomi <mimi@techaro.lol>",
      },
    ],
  ],
};
