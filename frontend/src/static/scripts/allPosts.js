//get info about posts

async function getPostData(){
    try{
        const response = await fetch(`http://localhost:8080/posts`);
        if (!response.ok){
            throw new Error(`Error HTTP: ${response.status}`); 
        };
        const postData = await response.json();
        console.log(postData);
        renderAllPosts(postData);
    } catch(error){
        console.log("the error was catched", error);
    }
};
getPostData();
//function for rendering all posts
const postsContainer = document.querySelector(".posts");
function renderAllPosts(data){
  for (let i=0; i<data.posts.length; i++){
    const postID = data.posts[i].id;
    const userName = data.posts[i].author.username;
    const date = formatDate(data.posts[i].created_at);
    const title = data.posts[i].title;
    const category = data.posts[i].category_id;
    const likes = data.posts[i].likes;
    const amountComments = data.posts[i].commentIDs.length;
    const newEl = `
                    <a href="postInside.html?id=${postID}" aria-label="link to the post" class="post">
                    <div class="author">
                        <div class="image-container">
                            <img class="author_img" src="/src/static/imgs/default.jpg" alt="photo of ${userName}">
                        </div>
                        <div class="author-info">
                            <p class="username">${userName}</p>
                            <p class="date">${date}</p>
                        </div>
                    </div>
                    <h3 class="post_title">${title}</h3>
                    <div class="post-footer">
                        <div class="tags">
                            <div class="tags_item">${category}</div>
                        </div>
                        <div class="stats">
                            <p class="post_comments">${amountComments}</p>
                            <p class="post_likes">${likes}</p>
                        </div>
                    </div>
                </a>`;
                postsContainer.insertAdjacentHTML('beforeend', newEl);
  };
};
//formating of the date
function formatDate(isoString){
    const date = new Date(isoString);
    return date.toLocaleDateString(undefined, { day: 'numeric', month: 'long', year: 'numeric' });
};
