import { useErrorResponseData } from '@nao-todo-server/hooks';
import multer, { MulterError } from 'multer';
import type { Request, Response } from 'express';

const upload = multer({
    storage: multer.memoryStorage(),
    limits: { fileSize: 1024 * 1024 * 2 },
    fileFilter: (req, file, cb) => {
        if (file.mimetype.startsWith('image/')) {
            cb(null, true);
        } else {
            cb(new Error('文件类型错误，请上传图片文件'));
        }
    }
});

const handleMulterError = (
    err: MulterError | Error,
    req: Request,
    res: Response
) => {
    console.error(`[api/user/uploadAvatar] Error: ${err.message}`);
    res.json(useErrorResponseData('头像上传失败，请稍候重试'));
};

export { upload, handleMulterError };
