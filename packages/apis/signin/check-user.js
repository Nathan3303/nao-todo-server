const md5 = require("md5");
const { User } = require("../../models");

module.exports = async function checkUser(email, password) {
    return await User.findOne({
        email,
        password: md5(password),
    }).select({
        id: { $toString: "$_id" },
        email: 1,
        nickName: 1,
        // _id: 0,
    });
};
