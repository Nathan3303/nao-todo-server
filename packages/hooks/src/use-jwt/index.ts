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
const useJWT = (payload: any, options: UseJWTOptions = defaultOptions) => {
    const { secret, header } = options;

    // 将header和payload转换为字符串
    const headerString = JSON.stringify(header);
    const payloadString = JSON.stringify(payload);

    // 将header和payload转换为base64编码
    const jwtHeader = Buffer.from(headerString).toString('base64');
    const jwtPayload = Buffer.from(payloadString).toString('base64');
    // 生成签名
    const jwtSignature = sign(`${jwtHeader}.${jwtPayload}`, secret);

    // 返回JWT
    return `${jwtHeader}.${jwtPayload}.${jwtSignature}`;
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

const getJWTPayload = (jwt: string) => {
    // 移除 Bearer 前缀
    jwt = jwt.replace('Bearer ', '');
    // 截取JWT的payload部分
    const [_, jwtPayload] = jwt.split('.');
    // 将payload转换为对象
    const payload = JSON.parse(Buffer.from(jwtPayload, 'base64').toString());
    // 返回payload
    return payload;
};

// 导出useJWT和verifyJWT函数
export { useJWT, verifyJWT, getJWTPayload };
