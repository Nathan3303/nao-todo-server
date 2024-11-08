import nodeResolve from '@rollup/plugin-node-resolve';
// import typescript2 from 'rollup-plugin-typescript2';
import typescript from '@rollup/plugin-typescript';
import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
import { defineConfig } from 'rollup';

export default defineConfig({
    input: 'src/index.ts', // 入口文件
    output: {
        dir: 'prod-dist', // 输出目录
        format: 'esm' // 输出格式（ES Module）
    },
    plugins: [
        nodeResolve({
            extensions: ['.ts', '.tsx', '.js', '.json'] // 告诉 Rollup 要解析的文件扩展名
        }),
        // typescript2({
        //     tsconfig: './tsconfig.json' // 指定 TypeScript 配置文件的位置
        // })
        typescript({
            // 指定 TypeScript 配置文件的位置
            extends: '../../tsconfig.base.json',
            compilerOptions: {
                lib: ['ES2015', 'ES2020', 'DOM'],
                strict: true,
                skipLibCheck: true,
                forceConsistentCasingInFileNames: true,
                moduleResolution: 'node',
                resolveJsonModule: true,
                isolatedModules: true,
                outDir: './dist/production',
                rootDir: '../../'
            },
            include: ['src/**/*', '../../packages/**/*']
        }),
        commonjs(), // 将 CommonJS 模块转换为 ES Module
        json() // 处理 JSON 文件
    ]
});
