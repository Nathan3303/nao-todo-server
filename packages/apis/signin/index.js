const { buildRD, checkMethod } = require("../../utils");
const checkUser = require("./check-user");
const checkSession = require("./check-session");

module.exports = async (request, response) => {
    if (checkMethod(request, response, "POST")) return;
    const { email, password } = request.body;
    if (!email || !password) {
        response.status(200).json(buildRD.error("邮箱或密码不能为空"));
        return;
    }
    try {
        const user = await checkUser(email, password);
        if (!user) {
            response.status(200).json(buildRD.error("邮箱或密码错误"));
            return;
        }
        const token = await checkSession(request, user);
        if (!token) {
            response.status(200).json(buildRD.error("登录失败，请稍后重试"));
            return;
        }
        response.status(200).json(buildRD.success(token));
    } catch (error) {
        response.status(200).json(buildRD.error("登录失败，请稍后重试"));
    }
};
