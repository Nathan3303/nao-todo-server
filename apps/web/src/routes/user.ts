import express from 'express';
import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';
import { updateUserNickname } from '@nao-todo-server/apis';

const exRouter: express.Router = express.Router();

exRouter.post('/user/nickname', async (req, res) => {
    console.log('[/api/user/nickname] Route matched!');
    await connectAndRun(async () => {
        await updateUserNickname(req, res);
    });
});

export default exRouter;
