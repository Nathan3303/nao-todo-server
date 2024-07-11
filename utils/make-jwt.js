const crypto = require("crypto");

const CRYPTO_SECRET = "NAO_SERVER_JWT_SECRET";

function sign(info, key) {
    return crypto.createHmac("sha256", key).update(info).digest("hex");
}

function toSign(header, payload) {
    return sign(header + "." + payload, CRYPTO_SECRET);
}

function makeJWT(payload) {
    const jwt_header = btoa({ alg: "HS256", typ: "JWT" });
    const jwt_payload = btoa(JSON.stringify(payload));
    // const jwt_signature = sign(jwt_header + "." + jwt_payload, CRYPTO_SECRET);
    const jwt_signature = toSign(jwt_header, jwt_payload);
    const jwt = jwt_header + "." + jwt_payload + "." + jwt_signature;
    return {
        token: jwt,
        header: jwt_header,
        payload: jwt_payload,
        signature: jwt_signature,
    };
}

function verifyJWT(token) {
    const [header, payload, signature] = token.split(".");
    const expected_signature = toSign(header, payload);
    return signature === expected_signature;
}

module.exports = { toSign, makeJWT, verifyJWT };
