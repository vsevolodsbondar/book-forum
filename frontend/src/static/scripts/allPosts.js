//get info about posts

async function getPostData(){
    try{
        const response = await fetch(`http://localhost:8080/posts`);
        if (!response.ok){
            throw new Error(`Error HTTP: ${response.status}`); 
        };
        const postData = await response.json();
        console.log(postData);
    } catch(error){
        console.log("the error was catched", error);
    }
};
getPostData();

