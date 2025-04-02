// utils/fileHash.js
import { createReadStream } from 'fs';
import { createHash } from 'crypto';

// 流式处理大文件（推荐）
export const calculateFileHash = (filePath: string) => {
    return new Promise((resolve, reject) => {
        const hash = createHash('sha256');
        const stream = createReadStream(filePath);
        stream.on('data', chunk => hash.update(chunk));
        stream.on('end', () => resolve(hash.digest('hex')));
        stream.on('error', err => reject(err));
    });
};

// 适用于小文件的同步方法
export const calculateFileHashSync = (buffer: Buffer) => {
    return createHash('sha256').update(buffer).digest('hex');
};
