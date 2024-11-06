const md5 = require("md5");
const validator = require("validator");
const { User } = require("../../models");
const { buildRD, checkMethod } = require("../../utils");

module.exports = async (request, response) => {
    if (checkMethod(request, response, "POST")) return;

    const { email, password } = request.body;

    if (!email) {
        response.status(200).json(buildRD.error("邮箱不能为空"));
        return;
    }
    if (!password) {
        response.status(200).json(buildRD.error("密码不能为空"));
        return;
    }

    try {
        if (!validator.isEmail(email)) {
            response.status(200).json(buildRD.error("邮箱格式不正确"));
            return;
        }
        const userExists = await User.findOne({ email });
        if (userExists) {
            response.status(200).json(buildRD.error("邮箱已被注册"));
            return;
        }
        const signUpResult = await User.create({
            username: email,
            password: md5(password),
            email: email,
            nickName: `${email.split("@")[0]}`,
        });
        console.log(signUpResult);
        response.status(200).json(buildRD("20000", "注册成功"));
    } catch (error) {
        response.status(200).json(buildRD.error("注册失败"));
    }
};
