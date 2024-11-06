const useResponseData = (code: number, message: string, data: unknown) => {
    code = code || 20000;
    message = message || 'OK';
    data = data || null;
    return { code, message, data };
};

const useSuccessfulResponseData = (data: unknown) => {
    return useResponseData(20000, 'OK', data);
};

const useErrorResponseData = (message: string) => {
    message = message || 'Internal Server Error';
    return useResponseData(50000, message, null);
};

export { useResponseData, useSuccessfulResponseData, useErrorResponseData };
