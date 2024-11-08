export type UseJWTOptions = {
    secret: string;
    header?: { alg: string; typ: string };
};
