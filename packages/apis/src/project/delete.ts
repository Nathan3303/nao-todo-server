import { Project } from '@nao-todo-server/models';
import {
    useSuccessfulResponseData,
    useErrorResponseData
} from '@nao-todo-server/hooks';
import { ObjectId } from '@nao-todo-server/utils';
import type { Request, Response } from 'express';

const deleteProject = async (req: Request, res: Response) => {
    try {
        if (!req.query.id) throw new Error('参数错误，请求无效');

        const deletedProject = await Project.findByIdAndDelete({
            _id: new ObjectId(req.query.id as string)
        });
        if (!deletedProject) throw new Error('项目不存在');

        return res.json(
            useSuccessfulResponseData({ projectId: deletedProject._id })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/project/deleteProject] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const deleteProjects = async (req: Request, res: Response) => {};

export { deleteProject, deleteProjects };
