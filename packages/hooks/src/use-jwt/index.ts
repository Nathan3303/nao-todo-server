// 导入crypto模块
import crypto from 'crypto';
// 导入jwtCryptoSecret常量
import { jwtCryptoSecret } from './constants';
// 导入UseJWTOptions类型
import type { UseJWTOptions } from './types';

// 定义sign函数，用于生成签名
const sign = (info: string, key: string) => {
    // 使用HMAC-SHA256算法生成签名
    return crypto.createHmac('sha256', key).update(info).digest('hex');
};

// 定义默认的JWT选项
const defaultOptions: UseJWTOptions = {
    secret: jwtCryptoSecret,
    header: {
        alg: 'HS256',
        typ: 'JWT'
    }
};

// 定义useJWT函数，用于生成JWT
const useJWT = (payload: unknown, options: UseJWTOptions = defaultOptions) => {
    const { secret, header } = options;
    try {
        // 将header和payload转换为字符串
        const headerString = JSON.stringify(header);
        const payloadString = JSON.stringify(payload);

        // 将header和payload转换为base64编码
        const jwtHeader = btoa(headerString);
        const jwtPayload = btoa(encodeURIComponent(payloadString));

        // 生成签名
        const jwtSignature = sign(`${jwtHeader}.${jwtPayload}`, secret);

        // 返回JWT
        return `${jwtHeader}.${jwtPayload}.${jwtSignature}`;
    } catch (error) {
        console.log('[useJWT] Error:', error);
    }
};

// 定义verifyJWT函数，用于验证JWT
const verifyJWT = (jwt: string, secret: string = jwtCryptoSecret) => {
    // 将JWT按照'.'分割成三部分：header、payload和signature
    const [jwtHeader, jwtPayload, jwtSignature] = jwt.split('.');
    // 生成签名
    const signature = sign(`${jwtHeader}.${jwtPayload}`, secret);
    // 返回签名是否一致
    return signature === jwtSignature;
};

// 定义getJWTPayload函数，用于获取JWT的payload
const getJWTPayload = (jwt: string) => {
    // 移除 Bearer 前缀
    jwt = jwt.replace('Bearer ', '');
    // 截取JWT的payload部分
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const [_, jwtPayload] = jwt.split('.');
    // 返回payload
    return JSON.parse(decodeURIComponent(atob(jwtPayload)));
};

// 导出useJWT和verifyJWT函数
export { useJWT, verifyJWT, getJWTPayload };
