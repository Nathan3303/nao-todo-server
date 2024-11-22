import nodeResolve from '@rollup/plugin-node-resolve';
import typescript from '@rollup/plugin-typescript';
import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
import { defineConfig } from 'rollup';

export default defineConfig({
    input: 'index.ts', // 入口文件
    output: {
        dir: 'dist', // 输出目录
        format: 'esm', // 输出格式（ES Module）
        sourcemap: true // 生成源码映射文件
    },
    external: ['mongoose', 'mongodb'],
    plugins: [
        // 将 CommonJS 模块转换为 ES Module
        commonjs(),
        // 处理 JSON 文件
        json(),
        // 指定 TypeScript 配置文件的位置
        typescript({ tsconfig: './tsconfig.json' }),
        // 告诉 Rollup 要解析的文件扩展名
        nodeResolve({ extensions: ['.ts', '.tsx', '.js', '.json'] })
    ]
});
