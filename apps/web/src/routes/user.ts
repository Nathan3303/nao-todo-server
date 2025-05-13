import express from 'express';
import {
    updateUserAvatar,
    updateUserNickname,
    updateUserPassword
} from '@nao-todo-server/apis';
import {
    // handleMulterError,
    upload
} from '@nao-todo-server/apis/src/user/upload-avatar';

const exRouter: express.Router = express.Router();

exRouter.post('/user/nickname', async (req, res) => {
    console.log('[/api/user/nickname] Route matched!');
    // await connectAndRun(async () => {
    await updateUserNickname(req, res);
    // });
});

exRouter.post('/user/password', async (req, res) => {
    console.log('[/api/user/password] Route matched!');
    // await connectAndRun(async () => {
    await updateUserPassword(req, res);
    // });
});

exRouter.post('/user/avatar', upload.single('avatar'), async (req, res) => {
    console.log('[/api/user/avatar] Route matched!');
    await updateUserAvatar(req, res);
});

// exRouter.use(handleMulterError);

export default exRouter;
