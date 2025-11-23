export default {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "body-max-line-length": [2, "always", 99999],
    "footer-max-line-length": [2, "always", 99999],
    "signed-off-by": [2, "always"],
  },
};
