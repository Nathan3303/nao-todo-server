type Task = () => Array<never> | Promise<never>;

const serialExecute = async (tasks: Task[]) => {
    return tasks.reduce(async (prev, curr) => {
        return prev.then(async results => {
            try {
                const result = await curr();
                results = results.concat(result);
                return Promise.resolve(results);
            } catch (error) {
                return Promise.reject(error);
            }
        });
    }, Promise.resolve([]));
};

export default serialExecute;
