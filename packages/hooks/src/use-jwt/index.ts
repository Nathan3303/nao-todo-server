import crypto from 'crypto';
import { jwtCryptoSecret } from './constants';

const sign = (info: string, key: string) => {
    return crypto.createHmac('sha256', key).update(info).digest('hex');
};

const toSign = (header: string, payload: string) => {
    return sign(header + '.' + payload, jwtCryptoSecret);
};

const verify = (jwt: string) => {
    const [header, payload, signature] = jwt.split('.');
    const expected_signature = toSign(header, payload);
    return signature === expected_signature;
};

const useJWT = (payload: object) => {
    const jwtHeader = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
    const jwtPayload = btoa(JSON.stringify(payload));
    const jwtSignature = toSign(jwtHeader, jwtPayload);
    const jwt = jwtHeader + '.' + jwtPayload + '.' + jwtSignature;
    return {
        jwt,
        header: jwtHeader,
        payload: jwtPayload,
        signature: jwtSignature
    };
};

export { useJWT, verify as verifyJWT };
