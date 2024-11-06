import type { PipelineStage } from 'mongoose';

type Task = () => PipelineStage[];

export const serialExecute = async (tasks: Task[]) => {
    return tasks.reduce(async (prev, curr) => {
        return prev.then(async results => {
            try {
                const result = await curr();
                results = results.concat(result as any);
                return Promise.resolve(results);
            } catch (error) {
                return Promise.reject(error);
            }
        });
    }, Promise.resolve([]));
};
