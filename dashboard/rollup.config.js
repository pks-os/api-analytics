import svelte from "rollup-plugin-svelte";
import commonjs from "@rollup/plugin-commonjs";
import typescript from "@rollup/plugin-typescript";
import resolve from "@rollup/plugin-node-resolve";
import livereload from "rollup-plugin-livereload";
import terser from "@rollup/plugin-terser";
import preprocess from "svelte-preprocess";
import json from "@rollup/plugin-json";
import css from "rollup-plugin-css-only";

const production = !process.env.ROLLUP_WATCH;

export default [
  // Browser bundle
  {
    input: "src/main.ts",
    output: {
      sourcemap: true,
      format: "iife",
      name: "app",
      file: "public/bundle.js",
    },
    plugins: [
      css({
        // Optional: filename to write all styles to
        output: "bundle.css",
      }),
      svelte({
        preprocess: preprocess({ sourceMap: !production }),
        compilerOptions: {
          dev: !production,
          hydratable: true,
        }
        // css: (css) => {
        //     css.write("bundle.css");
        //   },
        // }),
      }),
      typescript({
        // sourceMap: !production,
        // inlineSources: !production,
      }),
      json(),
      resolve(),
      commonjs(),
      // App.js will be built after bundle.js, so we only need to watch that.
      // By setting a small delay the Node server has a chance to restart before reloading.
      !production &&
        livereload({
          watch: "public/App.js",
          delay: 300,
        }),
      production && terser(),
    ],
  },
  // Server bundle
  {
    input: "src/App.svelte",
    output: {
      exports: "default",
      sourcemap: false,
      // format: "es",
      name: "app",
      file: "public/App.js",
    },
    plugins: [
      css(),
      svelte({
        preprocess: preprocess({ sourceMap: !production }),
        compilerOptions: {
          generate: "ssr",
        }
      }),
      typescript({ sourceMap: !production }),
      json(),
      resolve(),
      commonjs(),
      production && terser(),
    ],
  },
];
