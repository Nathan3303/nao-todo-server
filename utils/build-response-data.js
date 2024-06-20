function buildResponseData(code, message, data) {
    code = code || "20000";
    message = message || "OK";
    data = data || null;
    return { code, message, data };
}

buildResponseData.success = function (data) {
    return buildResponseData("20000", "OK", data);
};

buildResponseData.error = function (message) {
    message = message || "Internal Server Error";
    return buildResponseData("50000", message, null);
};

module.exports = buildResponseData;
