const posts = await getPosts()

const forum = document.getElementById("forum")

posts.forEach(post => {
  const element = document.createElement("div")

  element.textContent = post.title

  forum.appendChild(element)
})

async function getPosts() {
  const response = await fetch("/api/forum/posts")

  return response.json()
}
