["amd64", "arm64"].forEach(goarch =>
    [deb, rpm].forEach(method => method.build({
        name: "yeet",
        description: "Yeet out scripts with maximum haste!",
        homepage: "https://techaro.lol",
        license: "CC0",
        goarch,

        build: (out) => {
            go.build("-o", `${out}/usr/bin/yeet`, "./cmd/yeet");
        },
    }))
);