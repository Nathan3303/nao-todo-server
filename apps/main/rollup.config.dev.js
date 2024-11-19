import nodeResolve from '@rollup/plugin-node-resolve';
import typescript from '@rollup/plugin-typescript';
import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
import { defineConfig } from 'rollup';
// import copy from 'rollup-plugin-copy';

export default defineConfig({
    input: 'src/development.ts', // 入口文件
    external: ['mongoose', 'mongodb'],
    output: {
        dir: 'dist/development', // 输出目录
        format: 'esm', // 输出格式（ES Module）
        sourcemap: true // 生成源码映射文件
    },
    plugins: [
        // 将 CommonJS 模块转换为 ES Module
        commonjs(),
        // 处理 JSON 文件
        json(),
        // 指定 TypeScript 配置文件的位置
        typescript({ tsconfig: './tsconfig.dev.json' }),
        // 告诉 Rollup 要解析的文件扩展名
        nodeResolve({ extensions: ['.ts', '.tsx', '.js', '.json'] }),
        // 复制静态资源文件
        // copy({
        //     targets: [{ src: 'src/public', dest: 'dist/development' }]
        // })
    ]
});
