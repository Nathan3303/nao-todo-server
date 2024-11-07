import nodeResolve from '@rollup/plugin-node-resolve';
// import typescript2 from 'rollup-plugin-typescript2';
import typescript from '@rollup/plugin-typescript';
import { defineConfig } from 'rollup';

export default defineConfig({
    input: 'src/index.ts', // 入口文件
    output: {
        dir: 'dist', // 输出目录
        format: 'esm' // 输出格式（ES Module）
    },
    plugins: [
        nodeResolve({
            extensions: ['.ts', '.tsx', '.js', '.json'] // 告诉 Rollup 要解析的文件扩展名
        }),
        // typescript2({
        //     tsconfig: './tsconfig.json' // 指定 TypeScript 配置文件的位置
        // })
        typescript({})
    ]
});
