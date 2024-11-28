import express from 'express';
// import connectAndRun from '@nao-todo-server/utils/src/connect-and-run';
import { updateUserNickname, updateUserPassword } from '@nao-todo-server/apis';

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

export default exRouter;
