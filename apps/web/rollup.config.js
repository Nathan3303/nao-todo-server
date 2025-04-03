import { defineConfig } from 'rollup';
import nodeResolve from '@rollup/plugin-node-resolve';
import typescript from '@rollup/plugin-typescript';
import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
import terser from '@rollup/plugin-terser';

// eslint-disable-next-line no-undef
const isProd = process.env.NODE_ENV === 'production';
// eslint-disable-next-line no-undef
const isDev = process.env.NODE_ENV === 'development';

export default defineConfig({
    input: 'src/index.ts',
    external: ['mongoose', 'mongodb', '@typegoose/typegoose', 'openai', 'multer'],
    output: {
        dir: 'dist',
        format: 'esm',
        sourcemap: isDev
    },
    plugins: [
        commonjs(),
        json(),
        typescript({ tsconfig: 'tsconfig.json' }),
        nodeResolve({ extensions: ['.ts', '.js', '.json'], preferBuiltins: true }),
        terser({
            compress: {
                drop_console: isProd && ['log'],
                drop_debugger: true,
                global_defs: {
                    '@PROD': JSON.stringify(isProd),
                    '@DEV': JSON.stringify(isDev)
                }
            },
            format: {
                semicolons: false,
                shorthand: true
            },
            mangle: {
                toplevel: true,
                eval: true,
                keep_classnames: true,
                keep_fnames: true
            }
        })
    ]
});
