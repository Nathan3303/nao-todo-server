const { buildRD, checkMethod } = require("../../utils");
const checkUser = require("./check-user");
const checkSession = require("./check-session");

module.exports = async (request, response) => {
    if (checkMethod(request, response, "POST")) return;
    const { email, password } = request.body;
    if (!email || !password) {
        response
            .status(200)
            .json(buildRD.error("Email and password are required."));
        return;
    }
    try {
        const user = await checkUser(email, password);
        if (!user) {
            response
                .status(200)
                .json(buildRD.error("Invalid email or password."));
            return;
        }
        const token = await checkSession(request, user);
        if (!token) {
            response
                .status(200)
                .json(buildRD.error("Failed to create session."));
            return;
        }
        response.status(200).json(buildRD.success(token));
    } catch (error) {
        response.status(200).json(buildRD.error(error.message));
    }
};
