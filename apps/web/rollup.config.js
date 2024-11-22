import nodeResolve from '@rollup/plugin-node-resolve';
import typescript from '@rollup/plugin-typescript';
import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
import { defineConfig } from 'rollup';
import terser from '@rollup/plugin-terser';

// eslint-disable-next-line no-undef
const isProd = process.env.NODE_ENV === 'production';
// eslint-disable-next-line no-undef
const isDev = process.env.NODE_ENV === 'development';

export default defineConfig({
    input: 'index.ts', // 入口文件
    external: isDev ? ['mongoose', 'mongodb', '@typegoose/typegoose'] : [],
    output: {
        dir: 'dist', // 输出目录
        format: 'cjs', // 输出格式（ES Module）
        sourcemap: false // 生成源码映射文件
    },
    plugins: [
        // 将 CommonJS 模块转换为 ES Module
        commonjs(),
        // 处理 JSON 文件
        json(),
        // 指定 TypeScript 配置文件的位置
        typescript({ tsconfig: 'tsconfig.json' }),
        // 告诉 Rollup 要解析的文件扩展名
        nodeResolve({ extensions: ['.ts', '.tsx', '.js', '.json'] }),
        // 压缩代码
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
