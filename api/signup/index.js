const md5 = require("md5");
const validator = require("validator");
const { User } = require("../../models");
const { buildRD, checkMethod } = require("../../utils");

module.exports = async (request, response) => {
    if (checkMethod(request, response, "POST")) return;

    const { email, password } = request.body;
    
    if (!email) {
        response.status(200).json(buildRD.error("Email is required"));
        return;
    }
    if (!password) {
        response.status(200).json(buildRD.error("Password is required"));
        return;
    }

    try {
        if (!validator.isEmail(email)) {
            response.status(200).json(buildRD.error("Invalid email"));
            return;
        }
        const userExists = await User.findOne({ email });
        if (userExists) {
            response.status(200).json(buildRD.error("User already exists"));
            return;
        }
        const signUpResult = await User.create({
            username: email,
            password: md5(password),
            email: email,
            nickName: `${email.split("@")[0]}`,
        });
        console.log(signUpResult);
        response
            .status(200)
            .json(buildRD("20000", "User created successfully"));
    } catch (error) {
        response.status(200).json(buildRD.error("Error creating user"));
    }
};
