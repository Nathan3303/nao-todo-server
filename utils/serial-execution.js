module.exports = async (tasks) => {
    return tasks.reduce((prev, curr) => {
        return prev.then(async (results) => {
            try {
                const result = await curr();
                results.push(result);
                return Promise.resolve(results);
            } catch (error) {
                results.push(null);
                return Promise.reject(error);
            }
        });
    }, Promise.resolve([]));
};
