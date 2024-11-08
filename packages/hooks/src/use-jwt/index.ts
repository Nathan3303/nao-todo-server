import crypto from 'crypto';
import { jwtCryptoSecret } from './constants';
import type { UseJWTOptions } from './types';

const sign = (info: string, key: string) => {
    return crypto.createHmac('sha256', key).update(info).digest('hex');
};

const defaultOptions: UseJWTOptions = {
    secret: jwtCryptoSecret,
    header: {
        alg: 'HS256',
        typ: 'JWT'
    }
};

const useJWT = (payload: any, options: UseJWTOptions = defaultOptions) => {
    const { secret, header } = options;

    const headerString = JSON.stringify(header);
    const payloadString = JSON.stringify(payload);
    // console.log(headerString, payloadString);

    const jwtHeader = Buffer.from(headerString).toString('base64');
    const jwtPayload = Buffer.from(payloadString).toString('base64');
    const jwtSignature = sign(`${jwtHeader}.${jwtPayload}`, secret);

    return `${jwtHeader}.${jwtPayload}.${jwtSignature}`;
};

export { useJWT };
