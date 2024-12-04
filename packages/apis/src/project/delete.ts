import { Project, Todo } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData,
    getJWTPayload,
    useResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const deleteProject = async (req: Request, res: Response) => {
    try {
        // 获取请求头中的用户 ID
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        // 判断请求参数是否正确
        if (!userId || !req.query.projectId) {
            res.json(useErrorResponseData('参数错误，请求无效'));
        }

        // 删除清单
        const deleteResult = await Project.deleteOne({
            _id: new ObjectId(req.query.projectId as string),
            userId: new ObjectId(userId)
        }).exec();

        // 判断删除结果
        if (!deleteResult.deletedCount) {
            res.json(useErrorResponseData('删除失败'));
        }

        // 删除清单下的所有待办
        const deleteTodosResult = await Todo.deleteMany({
            projectId: new ObjectId(req.query.projectId as string),
            userId: new ObjectId(userId)
        });

        // 响应
        res.json(
            useSuccessfulResponseData({
                projectId: req.query.projectId as string,
                deletedTodosCount: deleteTodosResult.deletedCount
            })
        );
    } catch (e: unknown) {
        console.log('[api/project/deleteProject] Error:', e);
        res.json(useResponseData(50001, '删除失败，请稍后再试', null));
    }
};

const deleteProjects = async () => {};

export { deleteProject, deleteProjects };
