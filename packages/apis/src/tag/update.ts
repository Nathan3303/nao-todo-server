import { Tag } from '@nao-todo-server/models';
import { ObjectId } from '@nao-todo-server/utils';
import {
    getJWTPayload,
    useErrorResponseData,
    useSuccessfulResponseData
} from '@nao-todo-server/hooks';
import type { Request, Response } from 'express';

const updateTag = async (req: Request, res: Response) => {
    try {
        const userId = getJWTPayload(req.headers.authorization as string)
            .userId as string;

        if (!userId || !req.query.tagId) {
            throw new Error('参数错误，请求无效');
        }

        const tagId = req.query.tagId as string;

        const updatedTag = await Tag.findOneAndUpdate(
            { _id: new ObjectId(tagId), userId },
            { $set: { ...req.body } },
            { new: true }
        );

        if (!updatedTag) throw new Error('更新失败');

        return res.json(
            useSuccessfulResponseData({ tagId: updatedTag._id.toString() })
        );
    } catch (e: unknown) {
        if (e instanceof Error) {
            return res.json(useErrorResponseData(e.message));
        }
        console.log('[api/tag/updateTag] Error:', e);
        return res.json(useErrorResponseData('服务器错误'));
    }
};

const updateTags = async (req: Request, res: Response) => {};

export { updateTag, updateTags };
